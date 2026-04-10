#!/bin/bash

true_size() {
    wc -c < "$1" | tr -d ' '
}

FILE_PATH="data/test.txt"
FILE_SIZE=$(true_size "$FILE_PATH")

# 1. Get the presigned URL and extract it with jq
# We use sed to replace 0.0.0.0 with localhost so curl can connect
rust_fs_url=$(curl -s -X POST \
  localhost:7678/upload/presign \
  -H "Content-Type: application/json" \
  -d "{
    \"filename\": \"test.txt\",
    \"content_type\": \"text/plain\",
    \"content_length\": $FILE_SIZE
  }" | jq -r '.upload_url' )

# 2. Check if we actually got a URL
if [ -z "$rust_fs_url" ] || [ "$rust_fs_url" == "null" ]; then
    echo "Error: Could not retrieve upload URL"
    exit 1
fi

echo $rust_fs_url
# 3. Perform the actual upload
curl -X PUT "$rust_fs_url" \
     -H "Content-Type: text/plain" \
     -H "Content-Length: $FILE_SIZE" \
     --data-binary "@$FILE_PATH"
