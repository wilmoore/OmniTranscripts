#!/bin/bash
# Poll for async job completion
#
# Usage: ./async-job-polling.sh <job_id>
# Example: ./async-job-polling.sh 123e4567-e89b-12d3-a456-426614174000

set -e

JOB_ID="${1:?Usage: ./async-job-polling.sh <job_id>}"
API_KEY="${API_KEY:-your-api-key-here}"
BASE_URL="${BASE_URL:-http://localhost:3000}"
POLL_INTERVAL="${POLL_INTERVAL:-5}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-60}"

echo "Polling job: $JOB_ID"
echo "Poll interval: ${POLL_INTERVAL}s, Max attempts: $MAX_ATTEMPTS"

attempt=1
while [[ $attempt -le $MAX_ATTEMPTS ]]; do
  echo -n "Attempt $attempt: "

  response=$(curl -s -X GET "$BASE_URL/transcribe/$JOB_ID" \
    -H "Authorization: Bearer $API_KEY")

  status=$(echo "$response" | jq -r '.status // empty')

  case "$status" in
    "complete")
      echo "Complete!"
      echo ""
      echo "--- Result ---"
      echo "$response" | jq .
      exit 0
      ;;
    "error")
      echo "Error!"
      echo ""
      echo "--- Error Details ---"
      echo "$response" | jq .
      exit 1
      ;;
    "pending"|"running")
      echo "$status"
      ;;
    *)
      echo "Unknown status: $status"
      echo "$response" | jq .
      exit 1
      ;;
  esac

  sleep "$POLL_INTERVAL"
  ((attempt++))
done

echo "Max attempts reached. Job may still be processing."
exit 1
