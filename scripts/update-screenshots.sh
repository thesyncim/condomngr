#!/bin/bash
set -e

# Variables
SERVER_PID=""
TIMEOUT=10
SERVER_PORT=8080
SCREENSHOT_DELAY=3
USE_SIMPLE_CAPTURE=true # Set to true to use the simple capture method instead of Puppeteer

# Stop any existing server processes on port 8080
stop_existing_servers() {
  local pid=$(lsof -t -i:${SERVER_PORT} 2>/dev/null)
  if [ -n "$pid" ]; then
    echo "Stopping existing process on port ${SERVER_PORT} (PID: $pid)..."
    kill -9 $pid 2>/dev/null || true
    sleep 1
  fi
}

# Always stop existing servers before starting
stop_existing_servers

# Function to cleanup on exit
cleanup() {
  echo "Cleaning up..."
  if [ -n "$SERVER_PID" ]; then
    echo "Stopping server (PID: $SERVER_PID)..."
    kill -9 $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
  fi
  echo "Done"
}

# Set up cleanup trap
trap cleanup EXIT INT TERM

# Ensure screenshots directory exists
mkdir -p docs/screenshots

# Create placeholder images (debug mode)
create_placeholders() {
  echo "Creating placeholder screenshots..."
  for page in "dashboard" "residents" "payments" "expenses" "reports"; do
    echo "Creating placeholder for $page"
    echo "DEBUG MODE: Screenshot of $page page - $(date)" > "docs/screenshots/$page.png"
  done
}

# Ensure we have a built binary
echo "Building application..."
go build -o condomngr .

# Double-check port is available
if lsof -i:${SERVER_PORT} > /dev/null 2>&1; then
  echo "Error: Port ${SERVER_PORT} is still in use after cleanup. Please stop the process manually."
  exit 1
fi

# Start the server in sample mode
echo "Starting server in sample mode..."
./condomngr -sample &
SERVER_PID=$!

# Wait for server to start
echo "Waiting $TIMEOUT seconds for server to start..."
for i in $(seq 1 $TIMEOUT); do
  if curl -s http://localhost:$SERVER_PORT &>/dev/null; then
    echo "Server is up and running"
    break
  fi
  
  # Check if process is still running
  if ! ps -p $SERVER_PID > /dev/null; then
    echo "Error: Server process died. Check logs for details."
    create_placeholders
    node scripts/update-readme.js
    exit 1
  fi
  
  if [ $i -eq $TIMEOUT ]; then
    echo "Error: Server failed to start in $TIMEOUT seconds"
    create_placeholders
    node scripts/update-readme.js
    exit 1
  fi
  
  echo "Waiting... ($i/$TIMEOUT)"
  sleep 1
done

# Give the server a bit more time to fully initialize
echo "Waiting an additional $SCREENSHOT_DELAY seconds before capturing screenshots..."
sleep $SCREENSHOT_DELAY

# Capture screenshots
echo "Capturing screenshots..."
if [ "$USE_SIMPLE_CAPTURE" = true ]; then
  # Use simple shell-based screenshot capture
  scripts/capture-simple.sh || {
    echo "Error: Failed to capture screenshots using simple method"
    create_placeholders
  }
else
  # Try to use Puppeteer if requested (not recommended)
  node scripts/capture-screenshots.js || {
    echo "Error: Failed to capture screenshots using Puppeteer"
    create_placeholders
  }
fi

# Update README
echo "Updating README with screenshots..."
node scripts/update-readme.js

echo "Screenshot update process completed successfully" 