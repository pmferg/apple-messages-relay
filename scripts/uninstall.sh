#!/bin/bash
set -e

PLIST_NAME="com.example.messagesrelay.plist"
LAUNCH_AGENTS="$HOME/Library/LaunchAgents"
PLIST_PATH="$LAUNCH_AGENTS/$PLIST_NAME"
BIN_PATH="$HOME/.local/bin/messages-relay"
APP_SUPPORT="$HOME/Library/Application Support/messages-relay"

echo "=== messages-relay Uninstaller ==="

# Unload LaunchAgent
if [ -f "$PLIST_PATH" ]; then
	echo "Unloading LaunchAgent..."
	launchctl bootout "gui/$(id -u)/com.example.messagesrelay" 2>/dev/null || true
	rm -f "$PLIST_PATH"
	echo "Removed $PLIST_PATH"
fi

# Remove binary
if [ -f "$BIN_PATH" ]; then
	rm -f "$BIN_PATH"
	echo "Removed $BIN_PATH"
fi

# Optionally remove config and data
read -p "Remove config and Application Support data? [y/N] " REMOVE_DATA
if [ "$REMOVE_DATA" = "y" ] || [ "$REMOVE_DATA" = "Y" ]; then
	rm -rf "$APP_SUPPORT"
	echo "Removed $APP_SUPPORT"
fi

echo ""
echo "Uninstall complete."
