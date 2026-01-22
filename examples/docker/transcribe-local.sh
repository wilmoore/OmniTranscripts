#!/bin/bash
# Transcribe a local file using OmniTranscripts in Docker
#
# Note: The file must be in the mounted ./media directory
#
# Usage: ./transcribe-local.sh <filename>
# Example: ./transcribe-local.sh video.mp4

set -e

FILENAME="${1:?Usage: ./transcribe-local.sh <filename>}"
API_KEY="${API_KEY:-your-api-key-here}"
BASE_URL="${BASE_URL:-http://localhost:3000}"

MEDIA_DIR="$(dirname "$0")/media"

if [[ ! -f "$MEDIA_DIR/$FILENAME" ]]; then
  echo "Error: File not found: $MEDIA_DIR/$FILENAME"
  echo "Note: Place files in the ./media directory for Docker access"
  exit 1
fi

echo "Transcribing local file: $FILENAME"

curl -s -X POST "$BASE_URL/transcribe" \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@$MEDIA_DIR/$FILENAME" | jq .
