#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_NAME="messages-relay"
BIN_DIR="$HOME/.local/bin"
APP_SUPPORT="$HOME/Library/Application Support/messages-relay"
LOG_DIR="$HOME/Library/Logs/messages-relay"
LAUNCH_AGENTS="$HOME/Library/LaunchAgents"
PLIST_NAME="com.example.messagesrelay.plist"
PLIST_PATH="$LAUNCH_AGENTS/$PLIST_NAME"

echo "=== messages-relay Installer ==="
echo ""

# Prompt for configuration
read -p "MQTT Broker (e.g. ssl://broker.example.com:8883): " MQTT_BROKER
read -p "MQTT Topic: " MQTT_TOPIC
read -p "MQTT Username: " MQTT_USER
read -s -p "MQTT Password: " MQTT_PASS
echo ""
read -s -p "Shared secret (for HMAC): " SHARED_SECRET
echo ""
read -p "Allowed destinations (comma-separated, empty = allow all): " ALLOWED_DEST

# Create directories
mkdir -p "$BIN_DIR"

# Build Go binary
echo ""
echo "Building binary..."
cd "$PROJECT_DIR"
go build -o "$BIN_DIR/$BIN_NAME" ./cmd/messages-relay
mkdir -p "$APP_SUPPORT"
mkdir -p "$LOG_DIR"
mkdir -p "$LAUNCH_AGENTS"

# Build allowed_destinations JSON array
ALLOWED_JSON="[]"
if [ -n "$ALLOWED_DEST" ]; then
	# Split by comma, wrap each in quotes, join as JSON array
	ALLOWED_JSON="["
	FIRST=1
	IFS=',' read -ra DESTS <<< "$ALLOWED_DEST"
	for d in "${DESTS[@]}"; do
		d=$(echo "$d" | xargs)
		[ -z "$d" ] && continue
		[ $FIRST -eq 0 ] && ALLOWED_JSON+=", "
		ALLOWED_JSON+="\"$d\""
		FIRST=0
	done
	ALLOWED_JSON+="]"
fi

# Escape config values for JSON (basic: escape backslash and quote)
escape_json() {
	printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}
MQTT_BROKER_E=$(escape_json "$MQTT_BROKER")
MQTT_TOPIC_E=$(escape_json "$MQTT_TOPIC")
MQTT_USER_E=$(escape_json "$MQTT_USER")
MQTT_PASS_E=$(escape_json "$MQTT_PASS")
SHARED_SECRET_E=$(escape_json "$SHARED_SECRET")

# Write config.json
cat > "$APP_SUPPORT/config.json" << EOF
{
  "mqtt": {
    "broker": "$MQTT_BROKER_E",
    "topic": "$MQTT_TOPIC_E",
    "username": "$MQTT_USER_E",
    "password": "$MQTT_PASS_E"
  },
  "security": {
    "shared_secret": "$SHARED_SECRET_E",
    "max_skew_seconds": 60,
    "allowed_destinations": $ALLOWED_JSON
  },
  "limits": {
    "max_per_minute": 5,
    "max_per_day": 50
  },
  "relay": {
    "applescript_path": "$APP_SUPPORT/send-message.applescript",
    "test_mode": false
  }
}
EOF

# Copy AppleScript
cp "$SCRIPT_DIR/send-message.applescript" "$APP_SUPPORT/send-message.applescript"

# Generate LaunchAgent plist
BINARY_PATH="$BIN_DIR/$BIN_NAME"
LOG_PATH="$LOG_DIR/messages-relay.log"
WORKING_DIR="$APP_SUPPORT"

sed -e "s|{{BINARY_PATH}}|$BINARY_PATH|g" \
    -e "s|{{LOG_PATH}}|$LOG_PATH|g" \
    -e "s|{{WORKING_DIR}}|$WORKING_DIR|g" \
    "$PROJECT_DIR/deploy/com.example.messagesrelay.plist.template" \
    > "$PLIST_PATH"

# Load LaunchAgent
echo ""
echo "Loading LaunchAgent..."
SVC_ID="gui/$(id -u)/com.example.messagesrelay"
launchctl bootout "$SVC_ID" 2>/dev/null || true  # Unload if already running (reinstall)
launchctl bootstrap "gui/$(id -u)" "$PLIST_PATH"
launchctl enable "$SVC_ID" 2>/dev/null || true
launchctl kickstart -k "$SVC_ID"

echo ""
echo "=== Installation complete ==="
echo ""
echo "Next steps:"
echo "1. Grant Messages.app permission when prompted (System Settings > Privacy & Security > Automation)"
echo "2. Ensure you are logged in to iMessage in Messages.app"
echo "3. Check logs: tail -f $LOG_PATH"
echo ""
echo "To uninstall: ./scripts/uninstall.sh"
