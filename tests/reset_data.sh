#!/bin/bash

docker exec metadata-server sh -c "rm -rf /app/db/*"
docker restart metadata-server
docker exec osd-server sh -c "rm -rf /app/store/*"