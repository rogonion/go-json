#!/bin/bash
# Usage: ./install_vscode.sh <COMMIT_ID>

COMMIT_ID=$1
if [ -z "$COMMIT_ID" ]; then
    echo "⚠️  No Commit ID provided. Skipping pre-install."
    exit 0 # Exit 0 so the build doesn't fail if we don't have an ID yet
fi

PLATFORM="linux-x64"
SERVER_DIR="$HOME/.vscode-server/bin/$COMMIT_ID"
TARBALL="vscode-server.tar.gz"
DOWNLOAD_URL="https://update.code.visualstudio.com/commit:$COMMIT_ID/server-$PLATFORM/stable"

echo "⬇️  Installing VS Code Server for commit: $COMMIT_ID"

# 1. Clean previous attempts
rm -rf "$SERVER_DIR"
mkdir -p "$SERVER_DIR"

# 2. Download
wget -q -O "$SERVER_DIR/$TARBALL" "$DOWNLOAD_URL"
if [ $? -ne 0 ]; then
    echo "❌ Download failed! Check network or Commit ID."
    exit 1
fi

# 3. Extract
tar -xzf "$SERVER_DIR/$TARBALL" -C "$SERVER_DIR" --strip-components 1
rm "$SERVER_DIR/$TARBALL"

# 4. Success Marker
touch "$SERVER_DIR/0"
echo "✅ VS Code Server installed manually."