#!/bin/bash
# Interactive script to generate a valid messages-relay test payload.
# Prompts for secret, destination, and message (message has optional default).

set -e

echo "=== messages-relay Test Payload Generator ==="
echo ""

read -s -p "Shared secret: " SECRET
echo ""
if [[ -z "$SECRET" ]]; then
	echo "Error: secret is required."
	exit 1
fi

read -p "Destination (E.164, e.g. +447700900123): " DEST
if [[ -z "$DEST" ]]; then
	echo "Error: destination is required."
	exit 1
fi

read -p "Message [Hello from MQTT]: " PAYLOAD
PAYLOAD="${PAYLOAD:-Hello from MQTT}"

TS=$(date +%s)
NONCE=$(uuidgen 2>/dev/null || echo "test-${TS}-$$")

CANONICAL="${DEST}
${PAYLOAD}
${TS}
${NONCE}"

# Escape payload for JSON (backslash and double-quote)
PAYLOAD_ESC=$(printf '%s' "$PAYLOAD" | sed 's/\\/\\\\/g; s/"/\\"/g')
DEST_ESC=$(printf '%s' "$DEST" | sed 's/\\/\\\\/g; s/"/\\"/g')

HASH=$(printf '%s' "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -binary | xxd -p -c 256)

echo ""
echo "Payload (valid for ~60 seconds):"
echo ""
echo '{"destination":"'"$DEST_ESC"'","payload":"'"$PAYLOAD_ESC"'","timestamp":'"$TS"',"nonce":"'"$NONCE"'","hash":"'"$HASH"'"}'
echo ""
