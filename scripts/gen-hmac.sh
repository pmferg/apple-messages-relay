#!/bin/bash
# Generate HMAC for a messages-relay payload.
# Usage: ./gen-hmac.sh <shared_secret> <destination> <payload> [timestamp] [nonce]
#
# Example:
#   ./gen-hmac.sh "my-secret" "+447700900123" "Hello" 1700000000 "uuid-123"

set -e

if [ $# -lt 3 ]; then
	echo "Usage: $0 <shared_secret> <destination> <payload> [timestamp] [nonce]"
	exit 1
fi

SECRET="$1"
DEST="$2"
PAYLOAD="$3"
TS="${4:-$(date +%s)}"
NONCE="${5:-$(uuidgen 2>/dev/null || echo "$(date +%s)-$$")}"

CANONICAL="${DEST}
${PAYLOAD}
${TS}
${NONCE}"

HASH=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" -binary | xxd -p -c 256)

echo '{"destination":"'"$DEST"'","payload":"'"$PAYLOAD"'","timestamp":'"$TS"',"nonce":"'"$NONCE"'","hash":"'"$HASH"'"}'
