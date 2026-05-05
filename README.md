# coop-storage — Metadata Server

A file metadata store built on BadgerDB (LSM-tree key-value store) with RustFS as the object storage backend. Built for DSCI 551, Spring '26.

## Prerequisites

- Docker + Docker Compose
- Go 1.25+
- Git Bash (Windows) or bash (Mac/Linux)

## Setup

### 1. Clone and enter the repo

```bash
git clone https://github.com/michaelyiu1/coop-storage.git
cd coop-storage
```

### 2. Start the containers

```bash
cd devops
docker compose -f docker-compose.dev.yml up -d
```

This starts:
- `metadata-server` on port `7678`
- `osd-server` (RustFS) on port `9000` (S3 API) and `9001` (web console)

### 3. Create the bucket

Open http://localhost:9001 in your browser and log in:
- Username: `rustfsadmin`
- Password: `rustfsadmin`

Create a bucket named **`tests`**.

### 4. Verify the server is healthy

```bash
curl http://localhost:7678/health
# expected: {"status":"ok"}
```

## Running the E2E Demo

Test images are in `tests/data/`. From the repo root:

```bash
cd tests
sh full_test.sh
```
Expected output:

metadata-server
osd-server
2026/05/05 11:57:41 Got presign URL for object_key: anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt
2026/05/05 11:57:41 File uploaded successfully. ObjectKey: anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt
Upload completed.
2026/05/05 11:57:41 Got download URL: http://127.0.0.1:9000/tests/anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=rustfsadmin%2F20260505%2F%2Fs3%2Faws4_request&X-Amz-Date=20260505T185741Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=58294dc17556915a53016ee9ddbf2067df9e20be5a778d8f79920c1413e5abd3
2026/05/05 11:57:41 File downloaded to ./data/downloaded_test.txt
Download completed.
Metadata written.
2026/05/05 11:57:41 Metadata: {"id":"anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt","owner":"","fileType":"image/jpeg","fileName":"test.txt","deleteFlag":false,"version":0}
Metadata read.