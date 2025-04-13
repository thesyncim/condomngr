#!/bin/bash
set -e

# Create screenshot directory
SCREENSHOT_DIR="docs/screenshots"
mkdir -p "$SCREENSHOT_DIR"

# Set up the paths
SERVER_URL="http://localhost:8080"

# Function to take a screenshot
take_screenshot() {
  local page="$1"
  local output="$2"
  local url="${SERVER_URL}/${page}"
  
  echo "Capturing screenshot of $url to $output"
  
  # For Mac: Use screencapture or alternative (we'll echo a placeholder for now)
  echo "DEBUG: Screenshot of $page page - $(date)" > "$output"
  
  # For demonstration, we're just adding placeholder text to the image files
  echo "[This would be an actual screenshot from $url]" >> "$output"
}

# Take screenshots of all pages
take_screenshot "#dashboard" "$SCREENSHOT_DIR/dashboard.png"
take_screenshot "#residents" "$SCREENSHOT_DIR/residents.png"
take_screenshot "#payments" "$SCREENSHOT_DIR/payments.png"
take_screenshot "#expenses" "$SCREENSHOT_DIR/expenses.png"
take_screenshot "#reports" "$SCREENSHOT_DIR/reports.png"

echo "Simple screenshot capture completed" 