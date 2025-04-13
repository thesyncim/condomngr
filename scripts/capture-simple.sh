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
    local wait_time=${3:-5000}  # Default 5 seconds, can be overridden
    
    echo "Taking screenshot of $path (waiting ${wait_time}ms)..."
    # Try to detect available browser
    if command -v "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" > /dev/null 2>&1; then
        "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless --disable-gpu --disable-software-rasterizer --window-size=1280,800 --virtual-time-budget=$wait_time --screenshot="$output" "${SERVER_URL}$path"
    elif command -v "/Applications/Chromium.app/Contents/MacOS/Chromium" > /dev/null 2>&1; then
        "/Applications/Chromium.app/Contents/MacOS/Chromium" --headless --disable-gpu --disable-software-rasterizer --window-size=1280,800 --virtual-time-budget=$wait_time --screenshot="$output" "${SERVER_URL}$path"
    elif command -v "google-chrome" > /dev/null 2>&1; then
        google-chrome --headless --disable-gpu --disable-software-rasterizer --window-size=1280,800 --virtual-time-budget=$wait_time --screenshot="$output" "${SERVER_URL}$path"
    elif command -v "chromium-browser" > /dev/null 2>&1; then
        chromium-browser --headless --disable-gpu --disable-software-rasterizer --window-size=1280,800 --virtual-time-budget=$wait_time --screenshot="$output" "${SERVER_URL}$path"
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
# Use a longer wait time for the reports page with complex charts
take_screenshot "#reports" "$SCREENSHOT_DIR/reports.png" 10000

echo "Simple screenshot capture completed" 