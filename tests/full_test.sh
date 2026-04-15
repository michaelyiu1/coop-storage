#!/bin/bash

docker exec metadata-server sh -c "rm -rf /app/db/*"
docker restart metadata-server

docker restart osd-server

sleep 2  # give containers time to fully start

go run e2e.go