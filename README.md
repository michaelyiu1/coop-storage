# README

```bash
.
├── cli-client      # a command line tool to test uploading files
├── devops          # Docker compose, TBD nginx stuff
├── metadata-server # metadata and auth endpoints 
└── osd-server      # Contains file data
```

To get started:
```bash
$ cd server
$ docker compose up -d --force-recreate --build 
```

# TODO:
## Security
- encrypt files
- User buckets (replicated across nodes?)
    - A user shall access a bucket only if it's theirs
- Auth flow:
    - Client pings Metadata server
    - Metadata server issues token, mappings etc
    - Client pings OSD server

## Architecture
- image preview as metadata (but maybe store those as objects too)
- rate limiting (with Nginx)
- distributed 

## Abtraction
- Abstract DB into objects
- use mux
- grpc and protobufs to share types?

## UX
- upload multiple files form (make super optimized)
- file explorer
- find how to stream multiple files at once?
- shareable collections (e.g., movies etc)

# WILO
- delete job
- download route
- Unique filename per user check, maybe need new index?

# Resources
- Series of articles on [replication](https://www.enjoyalgorithms.com/blog/storage-and-redundancy)
- https://github.com/google/go-cloud
- https://medium.com/@kamal.maiti/object-based-storage-architecture-b841e5842124
