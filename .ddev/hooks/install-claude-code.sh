#!/bin/bash
set -e

PIN="2.1.109"
CURRENT=$(claude --version 2>/dev/null | awk '{print $1}')

if [ "$CURRENT" = "$PIN" ]; then
    echo "claude-code already at pinned version $PIN"
    exit 0
fi

echo "claude-code $CURRENT → reinstalling pinned $PIN"
npm install -g "@anthropic-ai/claude-code@$PIN"
echo "claude-code now: $(claude --version)"
