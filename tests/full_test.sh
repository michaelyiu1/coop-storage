#!/bin/bash

docker exec metadata-server sh -c "rm -rf /app/db/*"
docker restart metadata-server

docker restart osd-server


# BUCKET_NAME="tests"
# ENDPOINT="http://localhost:9000"
# AWS_ACCESS_KEY_ID="rustfsadmin"
# AWS_SECRET_ACCESS_KEY="rustfsadmin"
# AWS_DEFAULT_REGION="us-east-1" # Often required even if local
# # Check if the bucket exists
# if aws s3api head-bucket --bucket "$BUCKET_NAME" --endpoint-url "$ENDPOINT" 2>/dev/null; then
#     echo "Bucket '$BUCKET_NAME' already exists."
# else
#     echo "Bucket '$BUCKET_NAME' does not exist. Creating..."
#     aws s3 mb "s3://$BUCKET_NAME" --endpoint-url "$ENDPOINT"
    
#     if [ $? -eq 0 ]; then
#         echo "Successfully created bucket '$BUCKET_NAME'."
#     else
#         echo "Failed to create bucket."
#         exit 1
#     fi
# fi

sleep 2  # give containers time to fully start

go run e2e.go
