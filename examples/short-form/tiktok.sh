#!/bin/bash
# Transcribe a TikTok video
#
# Usage: ./tiktok.sh [FULL_URL]
# Example: ./tiktok.sh "https://www.tiktok.com/@username/video/1234567890"

set -e

TIKTOK_URL="${1:-https://www.tiktok.com/@example/video/1234567890}"
API_KEY="${API_KEY:-your-api-key-here}"
BASE_URL="${BASE_URL:-http://localhost:3000}"

echo "Transcribing TikTok: $TIKTOK_URL"

curl -s -X POST "$BASE_URL/transcribe" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"$TIKTOK_URL\"}" | jq .
