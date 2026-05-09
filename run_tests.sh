#!/usr/bin/env bash
# Wrapper script to run tests with proper PATH on Windows Git Bash

# Kill any leftover processes from previous runs
kill $(lsof -ti:19876 2>/dev/null) 2>/dev/null || true

# Add jq to PATH
JQ_DIR="/c/Users/dell/AppData/Local/Microsoft/WinGet/Packages/jqlang.jq_Microsoft.Winget.Source_8wekyb3d8bbwe"

# Create python3 executable pointing to real Python
PYTHON_DIR="/c/Program Files/Python313"
mkdir -p /tmp/bin
cp "${PYTHON_DIR}/python.exe" /tmp/bin/python3.exe 2>/dev/null

# Set PATH with our tools FIRST (before Windows Store aliases)
export PATH="/tmp/bin:${JQ_DIR}:$PATH"

echo "Tools check:"
echo "  jq:      $(which jq 2>/dev/null || echo 'NOT FOUND')"
echo "  python3: $(which python3 2>/dev/null || echo 'NOT FOUND')"
echo "  curl:    $(which curl 2>/dev/null || echo 'NOT FOUND')"
python3 --version 2>&1
echo ""

# Run the test suite
cd /d/My_Projects/Proxy_Maze/proxy-maze
BASE_URL=http://localhost:8080 bash proxymaze_test.sh
