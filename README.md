# coop-storage

A file metadata store built on BadgerDB (LSM-tree key-value store) with RustFS as the object storage backend. Built for DSCI 551, Spring '26.

---

## Prerequisites

- [Docker + Docker Compose](https://docs.docker.com/get-docker/)
- Git Bash (Windows) or bash (Mac/Linux)

> **Note:** All Go dependencies are managed inside Docker — no local Go installation or `go mod download` is needed.

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

Open [http://localhost:9001](http://localhost:9001) in your browser and log in with:

- **Username:** `rustfsadmin`
- **Password:** `rustfsadmin`

Create a bucket named `tests`.

### 4. Verify the server is healthy

```bash
curl http://localhost:7678/health
# expected: {"status":"ok"}
```

---

## Dataset

Test files used by the E2E demo are pre-made and included in `tests/data/`. No external dataset download or data generation is required.

---

## Running the E2E Demo

From the repo root:

```bash
cd tests
sh full_test.sh
```

### Expected output

```
metadata-server osd-server
2026/05/05 11:57:41 Got presign URL for object_key: anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt
2026/05/05 11:57:41 File uploaded successfully. ObjectKey: anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt
Upload completed.
2026/05/05 11:57:41 Got download URL: http://127.0.0.1:9000/tests/anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt?...
2026/05/05 11:57:41 File downloaded to ./data/downloaded_test.txt
Download completed.
Metadata written.
2026/05/05 11:57:41 Metadata: {"id":"anonymous/a653d718-b2f8-4773-8add-af89aedfdc19/test.txt","owner":"","fileType":"image/jpeg","fileName":"test.txt","deleteFlag":false,"version":0}
Metadata read.
```