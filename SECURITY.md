# Security

This document describes the security design of messages-relay.

## Threat Model

messages-relay assumes:

- **Attacker can read MQTT traffic** — Use TLS for MQTT. Without TLS, payloads and credentials are visible.
- **Attacker may replay messages** — Mitigated by timestamp window and nonce cache.
- **Attacker may attempt forgery** — Mitigated by HMAC-SHA256 with a shared secret.
- **Config and binary are on a trusted machine** — Secrets in config.json are only as secure as the filesystem and user account.

## Authentication and Integrity

### HMAC-SHA256

Every message must include a `hash` field: the hex-encoded HMAC-SHA256 of a canonical string.

**Canonical string:**
```
destination + "\n" + payload + "\n" + timestamp + "\n" + nonce
```

- Order matters. Any change invalidates the hash.
- Both publisher and relay must share the same secret.
- Comparison uses constant-time equality to prevent timing attacks.

### Timestamp Window

- `timestamp` must be within ±`max_skew_seconds` of the relay's current time (default 60).
- Reduces replay window to a short period.

### Nonce

- Each `nonce` may only be used once.
- Stored in-memory and expired after ~2 minutes.
- Prevents reuse of valid messages even within the timestamp window.

## Rate Limiting

- **Per minute:** Configurable (default 5).
- **Per day:** Configurable (default 50).
- Protects against abuse and accidental loops.

## Destination Validation

- Destination must match E.164 format (e.g. `+447700900123`).
- Optional `allowed_destinations` whitelist restricts who can receive messages.
- Payload max length: 1000 characters.

## Secret Management

- **No secrets in code or repository.** All secrets come from `config.json`.
- Config location: `~/Library/Application Support/messages-relay/config.json`
- Ensure:
  - Config file permissions are restricted (e.g. 0600).
  - Only the running user can read the config.
  - Config is not committed to version control.

## TLS

- Use `ssl://` or `mqtts://` for MQTT when possible.
- TLS protects credentials and message content in transit.

## Reporting Issues

If you discover a security vulnerability, please report it responsibly. Do not open a public issue for security-sensitive findings.
