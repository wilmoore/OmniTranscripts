#!/bin/bash
# Transcribe a URL using OmniTranscripts in Docker
#
# Usage: ./transcribe-url.sh <url>
# Example: ./transcribe-url.sh "https://www.youtube.com/shorts/VIDEO_ID"

set -e

URL="${1:?Usage: ./transcribe-url.sh <url>}"
API_KEY="${API_KEY:-your-api-key-here}"
BASE_URL="${BASE_URL:-http://localhost:3000}"

echo "Transcribing URL: $URL"

curl -s -X POST "$BASE_URL/transcribe" \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"url\": \"$URL\"}" | jq .
