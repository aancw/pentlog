#!/bin/bash
# Test script for GIF generation with resolution selection

echo "Testing GIF generation with sample Kali session data..."
echo ""

# Test with direct file mode
echo "Test 1: Direct file conversion"
echo "This will prompt for resolution selection"
echo ""

# We'll use expect to automate the interactive prompts
# Select 1080p (option 1) and use default filename
expect -c '
spawn ./pentlog gif sample_data/session-kali-20260118-172621.tty
expect "Select Resolution:"
send "\r"
expect "Enter output filename"
send "\r"
expect eof
'

echo ""
echo "Test completed. Check the reports directory for the generated GIF."
echo ""

# Show the generated file
REPORTS_DIR="$HOME/.pentlog/reports"
if [ -d "$REPORTS_DIR" ]; then
    echo "Generated files:"
    ls -lh "$REPORTS_DIR"/*.gif 2>/dev/null || echo "No GIF files found"
fi
