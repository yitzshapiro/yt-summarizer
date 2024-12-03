#!/bin/bash

# Debug output
echo "Current directory: $(pwd)"
echo "Python version: $(python3 --version)"
echo "Virtual environment path: $(which python3)"

# Activate the virtual environment
source .venv/bin/activate

# More debug output
echo "After activation:"
echo "Python version: $(python3 --version)"
echo "Python path: $(which python3)"
echo "yt-dlp version: $(python3 -m yt_dlp --version)"

# Export the PATH to include the virtual environment's bin directory
export PATH="$PATH:/usr/local/bin:/usr/bin:/bin:$HOME/.local/bin:$(pwd)/.venv/bin"

# Verify yt-dlp is available
which yt-dlp
echo "Using yt-dlp from: $(which yt-dlp)"

# Start the Go server in the background
go run main.go &

# Store the Go server's PID
GO_PID=$!

# Change directory to yt-frontend and run pnpm dev
cd yt-frontend && pnpm dev

# When pnpm dev is terminated, kill the Go server
kill $GO_PID

# Deactivate the virtual environment when done
deactivate