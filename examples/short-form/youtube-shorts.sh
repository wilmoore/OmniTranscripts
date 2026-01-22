#!/bin/bash
# Transcribe a YouTube Short
#
# Usage: ./youtube-shorts.sh [VIDEO_ID]
# Example: ./youtube-shorts.sh dQw4w9WgXcQ

set -e

VIDEO_ID="${1:-dQw4w9WgXcQ}"
API_KEY="${API_KEY:-your-api-key-here}"
BASE_URL="${BASE_URL:-http://localhost:3000}"

echo "Transcribing YouTube Short: $VIDEO_ID"

curl -s -X POST "$BASE_URL/transcribe" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"https://www.youtube.com/shorts/$VIDEO_ID\"}" | jq .
