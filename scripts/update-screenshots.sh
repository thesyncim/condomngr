#!/bin/bash
set -e

# Variables
SERVER_PID=""
TIMEOUT=10
SERVER_PORT=8080
SCREENSHOT_DELAY=3
DEBUG_MODE=false # Set to false to use real screenshots with Puppeteer

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

if [ "$DEBUG_MODE" = true ]; then
  echo "Running in DEBUG MODE - Creating placeholder screenshots"
  
  # Create placeholder images for each page
  for page in "dashboard" "residents" "payments" "expenses" "reports"; do
    echo "Creating placeholder for $page"
    echo "DEBUG MODE: Screenshot of $page page - $(date)" > "docs/screenshots/$page.png"
  done
else
  # Normal mode - uses Puppeteer
  # Check if Node.js is installed
  if ! command -v node &> /dev/null; then
    echo "Error: Node.js is required but not installed"
    exit 1
  fi

  # Install puppeteer if not already installed
  if [ ! -d "node_modules/puppeteer" ]; then
    echo "Installing puppeteer..."
    npm install puppeteer
  fi

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
      exit 1
    fi
    
    if [ $i -eq $TIMEOUT ]; then
      echo "Error: Server failed to start in $TIMEOUT seconds"
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
  node scripts/capture-screenshots.js || {
    echo "Error: Failed to capture screenshots"
    exit 1
  }
fi

# Update README
echo "Updating README with screenshots..."
node scripts/update-readme.js

echo "Screenshot update process completed successfully" 