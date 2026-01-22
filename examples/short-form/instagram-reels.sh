#!/bin/bash
# Transcribe an Instagram Reel (public only)
#
# Usage: ./instagram-reels.sh [REEL_ID]
# Example: ./instagram-reels.sh CxampleReelID

set -e

REEL_ID="${1:-CxampleReelID}"
API_KEY="${API_KEY:-your-api-key-here}"
BASE_URL="${BASE_URL:-http://localhost:3000}"

echo "Transcribing Instagram Reel: $REEL_ID"
echo "Note: Only public reels are supported"

curl -s -X POST "$BASE_URL/transcribe" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"https://www.instagram.com/reel/$REEL_ID/\"}" | jq .
