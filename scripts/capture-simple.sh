#!/bin/bash
set -e

# Create screenshots directory if it doesn't exist
SCREENSHOT_DIR="screenshots"
mkdir -p "$SCREENSHOT_DIR"

# Server URL
SERVER_URL="http://localhost:8080"

# Function to take a screenshot
take_screenshot() {
    local path=$1
    local output=$2
    
    echo "Taking screenshot of $path..."
    # Try to detect available browser
    if command -v "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" > /dev/null 2>&1; then
        "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless --disable-gpu --window-size=1280,800 --virtual-time-budget=5000 --screenshot="$output" "${SERVER_URL}$path"
    elif command -v "/Applications/Chromium.app/Contents/MacOS/Chromium" > /dev/null 2>&1; then
        "/Applications/Chromium.app/Contents/MacOS/Chromium" --headless --disable-gpu --window-size=1280,800 --virtual-time-budget=5000 --screenshot="$output" "${SERVER_URL}$path"
    elif command -v "google-chrome" > /dev/null 2>&1; then
        google-chrome --headless --disable-gpu --window-size=1280,800 --virtual-time-budget=5000 --screenshot="$output" "${SERVER_URL}$path"
    elif command -v "chromium-browser" > /dev/null 2>&1; then
        chromium-browser --headless --disable-gpu --window-size=1280,800 --virtual-time-budget=5000 --screenshot="$output" "${SERVER_URL}$path"
    else
        echo "Error: No compatible browser found for taking screenshots."
        exit 1
    fi
    echo "Screenshot saved to $output"
}

# Take screenshots of all pages
take_screenshot "#dashboard" "$SCREENSHOT_DIR/dashboard.png"
take_screenshot "#residents" "$SCREENSHOT_DIR/residents.png"
take_screenshot "#payments" "$SCREENSHOT_DIR/payments.png"
take_screenshot "#expenses" "$SCREENSHOT_DIR/expenses.png"
take_screenshot "#reports" "$SCREENSHOT_DIR/reports.png"

echo "Simple screenshot capture completed" 