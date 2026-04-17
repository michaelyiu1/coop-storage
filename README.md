# coop-storage — Metadata Server

A file metadata store built on BadgerDB (LSM-tree key-value store) with RustFS as the object storage backend. Built for DSCI 551, Spring '26.

## Prerequisites

- Docker + Docker Compose
- Go 1.25+
- Git Bash (Windows) or bash (Mac/Linux)

## Setup

### 1. Clone and enter the repo

```bash
git clone <repo-url>
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

# expected: {"status":"ok"}
```

## Running the E2E Demo

Test images are in `tests/data/`. From the repo root:

```bash
cd tests
sh full_test.sh
```

Expected output: