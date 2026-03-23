# messages-relay

A production-quality personal automation tool that subscribes to an MQTT topic and relays validated messages via macOS Messages.app (iMessage).

**Intended use:** Personal automation only. Single-user, local-first, designed to run on a Mac mini or similar as a LaunchAgent.

---

## 1. Overview

### What it does

- Subscribes to a configurable MQTT topic
- Validates incoming JSON messages using HMAC-SHA256
- Prevents replay attacks (timestamp window + nonce cache)
- Sends approved messages via macOS Messages.app using AppleScript
- Runs as a LaunchAgent in the logged-in user session

### Intended use

- **Personal automation only.** Not suitable for bulk messaging, marketing, or multi-user systems.
- Run on your own Mac (e.g. Mac mini) with your own iMessage account.
- Trigger messages from home automation, monitoring systems, or custom scripts via MQTT.

---

## 2. Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   MQTT Broker   │────▶│  messages-relay  │────▶│  Messages.app   │
│  (TLS optional) │     │                  │     │  (iMessage)     │
└─────────────────┘     └────────┬─────────┘     └─────────────────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                    ▼            ▼            ▼
              ┌──────────┐ ┌──────────┐ ┌──────────┐
              │ Validator│ │ Security │ │  Relay   │
              │ (JSON)   │ │ (HMAC,   │ │(AppleScript)
              │          │ │  nonce,  │ │          │
              │          │ │  rate)   │ │          │
              └──────────┘ └──────────┘ └──────────┘
```

### Components

| Component | Responsibility |
|-----------|----------------|
| **MQTT Client** | Connects over TLS, subscribes to topic, auto-reconnects, forwards raw payloads |
| **Message Validator** | Parses JSON, validates required fields, E.164 destination, payload size |
| **Security** | HMAC-SHA256 verification, timestamp window (±60s), nonce replay cache, rate limiting |
| **Relay Engine** | Invokes `osascript` with AppleScript to send via Messages.app |
| **Config Loader** | Loads `config.json` from Application Support, no secrets in code |
| **Logging** | Structured logs to stdout + file, no full payloads logged by default |

---

## 3. Installation

### macOS requirements

- macOS with Messages.app (iMessage) configured
- User logged in (LaunchAgent runs in user session)
- Go 1.21+ (for building from source)
- Network access to your MQTT broker

### Step-by-step

1. **Clone or download** this repository.

2. **Run the installer:**
   ```bash
   ./scripts/install.sh
   ```

3. **Provide values when prompted:**
   - MQTT broker URL (e.g. `ssl://broker.example.com:8883`)
   - MQTT topic (e.g. `personal/messages/send`)
   - MQTT username and password
   - Shared secret (used for HMAC; use a long random string)
   - Optional: comma-separated allowed destinations (empty = allow all)

4. **Grant permissions** when macOS prompts:
   - System Settings → Privacy & Security → Automation → allow the terminal (or `messages-relay`) to control Messages

5. **Verify** Messages.app is signed in to iMessage and ready to send.

6. **Check logs** if needed:
   ```bash
   tail -f ~/Library/Logs/messages-relay/messages-relay.log
   ```

---

## 4. Configuration

Config file: `~/Library/Application Support/messages-relay/config.json`

| Field | Description |
|-------|-------------|
| `mqtt.broker` | Broker URL (`ssl://`, `tcp://`, `mqtts://`) |
| `mqtt.topic` | Topic to subscribe to |
| `mqtt.username` | MQTT username |
| `mqtt.password` | MQTT password |
| `security.shared_secret` | Secret for HMAC-SHA256 (keep private) |
| `security.max_skew_seconds` | Timestamp tolerance (default 60) |
| `security.allowed_destinations` | Optional whitelist; empty = allow all |
| `limits.max_per_minute` | Rate limit (default 5) |
| `limits.max_per_day` | Rate limit (default 50) |
| `relay.test_mode` | If true, no AppleScript (for testing) |

### Security notes

- Never commit `config.json` or share it.
- Use a strong, random `shared_secret` (e.g. 32+ bytes).
- Use TLS for MQTT in production (`ssl://` or `mqtts://`).
- Optionally restrict `allowed_destinations` to known recipients.

---

## 5. Usage

### Example MQTT message

Publish a JSON message to your configured topic:

```json
{
  "destination": "+447700900123",
  "payload": "Hello from MQTT",
  "timestamp": 1700000000,
  "nonce": "550e8400-e29b-41d4-a716-446655440000",
  "hash": "hex-hmac-sha256-of-canonical-string"
}
```

### HMAC generation

**Canonical string** (newline-separated, no trailing newline):

```
destination + "\n" + payload + "\n" + timestamp + "\n" + nonce
```

Example: `+447700900123\nHello\n1700000000\n550e8400-e29b-41d4-a716-446655440000`

**Go:**

```go
import "github.com/example/messages-relay/internal/security"

canonical := security.CanonicalInput(destination, payload, timestamp, nonce)
hash := security.ComputeHMAC(sharedSecret, canonical)
```

**Shell (OpenSSL):**

```bash
# canonical.txt = destination\npayload\ntimestamp\nnonce
openssl dgst -sha256 -hmac "$SHARED_SECRET" -binary canonical.txt | xxd -p -c 256
```

---

## 6. Security Model

### HMAC

- Algorithm: HMAC-SHA256
- Input: canonical string (destination, payload, timestamp, nonce)
- Comparison: constant-time to avoid timing attacks
- Shared secret must be kept confidential by both publisher and relay

### Replay protection

1. **Timestamp:** Message must be within ±60 seconds of server time.
2. **Nonce:** Each nonce may only be used once. Stored in-memory, expired after ~2 minutes.

### Why secrets are local-only

- No cloud storage of secrets
- Config lives in user Application Support
- MQTT credentials and shared secret never leave the machine

### Risks and limitations

- If the shared secret is compromised, attackers can forge valid messages.
- Nonce store is in-memory; restarting resets it (replay possible only within 2-minute window).
- Rate limits are per-process; multiple instances would have separate limits.
- Messages.app and AppleScript depend on macOS; failures may be opaque.

---

## 7. Limitations

- **Requires logged-in macOS user** — LaunchAgent runs in the user session
- **Requires Messages.app configured** — iMessage must be set up and signed in
- **macOS permission prompts** — Automation permission for Messages is required
- **Not suitable for bulk messaging** — Rate limits (5/min, 50/day) and iMessage limits apply
- **Single user** — No multi-tenant or shared deployment model
- **macOS only** — AppleScript and Messages.app are macOS-specific

---

## 8. Troubleshooting

### MQTT connection issues

- Verify broker URL, port, and TLS (`ssl://` vs `tcp://`)
- Check username/password
- Ensure firewall allows outbound MQTT (typically 8883 for TLS)
- Check logs: `tail -f ~/Library/Logs/messages-relay/messages-relay.log`

### AppleScript failures

- Confirm Messages.app is running and signed in
- Check Automation permissions (System Settings → Privacy & Security → Automation)
- Test manually: `osascript ~/Library/Application\ Support/messages-relay/send-message.applescript "+1234567890" "Test"`
- Ensure destination is in E.164 format (e.g. `+447700900123`)

### LaunchAgent debugging

```bash
# Check if loaded
launchctl print gui/$(id -u)/com.example.messagesrelay

# View recent logs
tail -100 ~/Library/Logs/messages-relay/messages-relay.log

# Restart
launchctl kickstart -k gui/$(id -u)/com.example.messagesrelay

# Unload (for testing)
launchctl bootout gui/$(id -u)/com.example.messagesrelay
```

---

## 9. Development

### Run locally without LaunchAgent

```bash
# Create config manually in ~/Library/Application Support/messages-relay/config.json
# Set relay.test_mode: true to avoid invoking AppleScript

go run ./cmd/messages-relay
```

### Test message validation

```bash
go test ./internal/validator/...
go test ./internal/security/...
```

### Mock relay for testing

Set `relay.test_mode: true` in config. The relay will accept messages but not invoke AppleScript.

---

## 10. License + Disclaimer

- **Personal-use tool** — Use at your own risk.
- **Not affiliated with Apple** — This project is independent of Apple Inc.
- **No guarantee of reliability** — Messages delivery depends on iMessage, MQTT broker, and macOS. No warranty is provided.
- **MIT License** — See LICENSE file if present.

---

## Uninstall

```bash
./scripts/uninstall.sh
```
