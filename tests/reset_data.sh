#!/bin/bash

docker exec metadata-server sh -c "rm -rf /app/db/*"
docker exec osd-server sh -c "rm -rf /app/store/*"