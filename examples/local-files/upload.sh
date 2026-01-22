#!/bin/bash
# Upload and transcribe a local file
#
# Usage: ./upload.sh <file_path>
# Example: ./upload.sh ./recording.mp4

set -e

FILE_PATH="${1:?Usage: ./upload.sh <file_path>}"
API_KEY="${API_KEY:-your-api-key-here}"
BASE_URL="${BASE_URL:-http://localhost:3000}"

if [[ ! -f "$FILE_PATH" ]]; then
  echo "Error: File not found: $FILE_PATH"
  exit 1
fi

echo "Uploading and transcribing: $FILE_PATH"

curl -s -X POST "$BASE_URL/transcribe" \
  -H "Authorization: Bearer $API_KEY" \
  -F "file=@$FILE_PATH" | jq .
