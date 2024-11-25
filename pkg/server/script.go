package server

const installScript = `#!/bin/bash
set -euo pipefail

# Usage function
usage() {
    echo "Usage: $0 -b BINARY_PATH -u DOWNLOAD_URL [-p PORT]"
    echo
    echo "Options:"
    echo "  -b  Path where the binary should be installed"
    echo "  -p  Port to run the server on (random if not specified)"
    echo "  -u  URL to download the binary from"
    exit 1
}

# Function to find a random free port
find_free_port() {
    # Try ss first (newer), fall back to netstat
    if command -v ss >/dev/null 2>&1; then
        while true; do
            # Generate random port between 1024-65535
            PORT=$(shuf -i 1024-65535 -n 1)
            if ! ss -ltn | grep -q ":$PORT\b"; then
                echo "$PORT"
                return 0
            fi
        done
    else
        while true; do
            PORT=$(shuf -i 1024-65535 -n 1)
            if ! netstat -tln | grep -q ":$PORT\b"; then
                echo "$PORT"
                return 0
            fi
        done
    fi
}

# Parse arguments
PORT=""
while getopts "b:p:u:" opt; do
    case $opt in
        b) BINARY_PATH="$OPTARG" ;;
        p) PORT="$OPTARG" ;;
        u) DOWNLOAD_URL="$OPTARG" ;;
        *) usage ;;
    esac
done

# Verify all required parameters are present
if [ -z "${BINARY_PATH:-}" ] || [ -z "${DOWNLOAD_URL:-}" ]; then
    usage
fi

# If port is not specified, find a free one
if [ -z "$PORT" ]; then
    PORT=$(find_free_port)
fi
echo "Using port: $PORT"

# Create directory if it doesn't exist
BINARY_DIR=$(dirname "$BINARY_PATH")
mkdir -p "$BINARY_DIR"

# Check if binary already exists and is running
if [ -f "$BINARY_PATH" ]; then
    echo "Binary already exists at $BINARY_PATH"
    if pgrep -f "$BINARY_PATH.*server run" >/dev/null; then
        echo "Server is already running"
        # Get the port from the running process
        RUNNING_PORT=$(ss -lptn | grep "$(pgrep -f "$BINARY_PATH.*server run")" | awk '{print $4}' | cut -d: -f2)
        echo "Port: $RUNNING_PORT"
        exit 0
    else
        echo "Existing binary found but server not running. Will reinstall."
    fi
fi

# Download the binary
echo "Downloading binary from $DOWNLOAD_URL..."
if ! curl -fsSL -o "$BINARY_PATH" "$DOWNLOAD_URL"; then
    echo "Failed to download binary"
    exit 1
fi

# Make binary executable
chmod +x "$BINARY_PATH"

# Verify binary works
if ! "$BINARY_PATH"; then
    echo "Failed to verify binary"
    rm -f "$BINARY_PATH"
    exit 1
fi

# Start server
echo "Starting server on port $PORT..."
LOG_FILE="${BINARY_DIR}/hide.log"
nohup "$BINARY_PATH" server run --workspace-dir "$(pwd)" --binary-dir "$HOME/.hide/bin" --port "$PORT" >"$LOG_FILE" 2>&1 &

# Wait for server to start (simplified)
MAX_ATTEMPTS=30
ATTEMPTS=0
while [ $ATTEMPTS -lt $MAX_ATTEMPTS ]; do
    if netstat -tln | grep -q ":$PORT\b" || ss -tln | grep -q ":$PORT\b"; then
        echo "Server started successfully on port $PORT"
        exit 0
    fi
    ATTEMPTS=$((ATTEMPTS + 1))
    sleep 2
done

echo "Server failed to start within 60 seconds"
exit 1
`
