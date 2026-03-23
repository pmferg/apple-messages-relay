#!/bin/bash
# Bootstrap installer: run via curl with no local dependencies.
# Installs Go if needed, downloads the repo, and runs the full installer.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/OWNER/REPO/main/scripts/bootstrap-install.sh | bash
#
# Or with a specific repo URL:
#   curl -fsSL BOOTSTRAP_URL | bash -s -- https://github.com/OWNER/REPO
#   curl -fsSL BOOTSTRAP_URL | REPO_URL=https://github.com/OWNER/REPO bash

set -e

# Default repo - override with first arg or REPO_URL env
DEFAULT_REPO="https://github.com/pmferg/apple-messages-relay.git"
REPO_URL="${REPO_URL:-${1:-$DEFAULT_REPO}}"
BRANCH="${BRANCH:-main}"

# Resolve owner/repo for GitHub archive URL
# Supports: https://github.com/owner/repo or https://github.com/owner/repo.git
if [[ "$REPO_URL" =~ ^https?://github\.com/([^/]+)/([^/]+?)(\.git)?$ ]]; then
	GITHUB_OWNER="${BASH_REMATCH[1]}"
	GITHUB_REPO="${BASH_REMATCH[2]%.git}"
	ARCHIVE_URL="https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/archive/refs/heads/${BRANCH}.zip"
else
	echo "Unsupported REPO_URL format. Use: https://github.com/pmferg/apple-messages-relay"
	exit 1
fi

echo "=== messages-relay Bootstrap Installer ==="
echo "Repository: $REPO_URL"
echo ""

# macOS only
if [[ "$(uname)" != "Darwin" ]]; then
	echo "Error: messages-relay requires macOS (Messages.app)."
	exit 1
fi

# Ensure required tools (curl, unzip are usually present on macOS)
for cmd in curl unzip; do
	if ! command -v "$cmd" &>/dev/null; then
		echo "Error: $cmd is required but not installed."
		exit 1
	fi
done

# Install Go if not present
ensure_go() {
	if command -v go &>/dev/null; then
		echo "Go found: $(go version)"
		return
	fi

	echo "Go not found. Installing..."
	if command -v brew &>/dev/null; then
		brew install go
		eval "$(brew shellenv 2>/dev/null)" || true
	elif [[ -x /opt/homebrew/bin/brew ]]; then
		/opt/homebrew/bin/brew install go
		export PATH="/opt/homebrew/bin:$PATH"
	else
		# Download Go directly
		ARCH=$(uname -m)
		case "$ARCH" in
			arm64) GOARCH="darwin-arm64" ;;
			x86_64) GOARCH="darwin-amd64" ;;
			*) echo "Unsupported architecture: $ARCH"; exit 1 ;;
		esac
		GO_VERSION="1.21.13"
		GO_TAR="go${GO_VERSION}.${GOARCH}.tar.gz"
		GO_URL="https://go.dev/dl/${GO_TAR}"
		GO_INSTALL="$HOME/.local/go-install"

		mkdir -p "$GO_INSTALL"
		curl -fsSL "$GO_URL" -o "$GO_INSTALL/$GO_TAR"
		tar -C "$GO_INSTALL" -xzf "$GO_INSTALL/$GO_TAR"
		export PATH="$GO_INSTALL/go/bin:$PATH"
		export GOROOT="$GO_INSTALL/go"
		echo "Installed Go to $GO_INSTALL"
	fi

	if ! command -v go &>/dev/null; then
		echo "Error: Failed to install Go. Please install manually: https://go.dev/dl/"
		exit 1
	fi
}

# Download repo and run installer
WORK_DIR=$(mktemp -d -t messages-relay-install.XXXXXX)
trap 'rm -rf "$WORK_DIR"' EXIT

echo "Downloading source..."
curl -fsSL "$ARCHIVE_URL" -o "$WORK_DIR/repo.zip"
unzip -q "$WORK_DIR/repo.zip" -d "$WORK_DIR"
REPO_DIR=$(find "$WORK_DIR" -maxdepth 1 -type d -name "*-${BRANCH}" 2>/dev/null | head -1)
if [[ -z "$REPO_DIR" ]]; then
	# Try "main" or "master" folder naming
	REPO_DIR=$(find "$WORK_DIR" -maxdepth 1 -type d ! -path "$WORK_DIR" 2>/dev/null | head -1)
fi
if [[ -z "$REPO_DIR" || ! -f "$REPO_DIR/scripts/install.sh" ]]; then
	echo "Error: Could not find repository structure."
	exit 1
fi

echo "Installing Go (if needed)..."
ensure_go

echo ""
"$REPO_DIR/scripts/install.sh"
