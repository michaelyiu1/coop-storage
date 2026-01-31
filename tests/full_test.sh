#!/bin/bash

# reset any data, and run end to end tests covering the osd and metadata servers
docker exec metadata-server sh -c "rm -rf /app/db/*"
docker restart metadata-server
docker exec osd-server sh -c "rm -rf /app/store/*"

sleep 0.2
go run e2e.go
