# ShortLink

A URL shortening service built in Go. Encodes long URLs into 6-character short codes and decodes them back, with persistence across restarts.

---

## DEMO

Check out [LINK DEMO](https://shortlink.gliam.me) for reaching the web.

https://github.com/user-attachments/assets/bde17b38-a48d-478a-b93b-37dc497c88b8

---

## Table of Contents

- [ShortLink](#shortlink)
  - [DEMO](#demo)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Tech Stack](#tech-stack)
  - [How to Run](#how-to-run)
    - [Prerequisites](#prerequisites)
    - [Option A: Docker Compose (recommended)](#option-a-docker-compose-recommended)
    - [Option B: Run locally](#option-b-run-locally)
  - [API Reference](#api-reference)
    - [POST /encode](#post-encode)
    - [POST /decode](#post-decode)
    - [GET /:code](#get-code)
  - [Running Tests](#running-tests)
    - [Load tests (k6)](#load-tests-k6)
  - [Security Considerations](#security-considerations)
    - [1. URL injection / SSRF](#1-url-injection--ssrf)
    - [2. Redirect abuse (open redirector)](#2-redirect-abuse-open-redirector)
    - [3. Enumeration / scraping](#3-enumeration--scraping)
    - [4. Denial of Service](#4-denial-of-service)
    - [5. SQL injection](#5-sql-injection)
    - [6. Logging sensitive data](#6-logging-sensitive-data)
    - [7. Header / host spoofing](#7-header--host-spoofing)
  - [Scalability \& Collision Strategy](#scalability--collision-strategy)
    - [Short code generation](#short-code-generation)
    - [Collision handling](#collision-handling)
    - [Horizontal scaling](#horizontal-scaling)
    - [Code space exhaustion](#code-space-exhaustion)
  - [My Tasks and AI Assignments](#my-tasks-and-ai-assignments)
    - [My Tasks](#my-tasks)
    - [AI Assignments](#ai-assignments)

---

## Features

- `POST /encode` — shorten any valid HTTP/HTTPS URL
- `POST /decode` — resolve a short URL back to the original
- `GET /:code` — browser-friendly redirect (HTTP 302)
- Idempotent encode: the same long URL always returns the same short code
- Persistent storage in PostgreSQL; short codes survive restarts
- Optional LRU or Redis cache for decode hot-path

---

## Tech Stack

| Concern    | Choice                           |
| ---------- | -------------------------------- |
| Language   | Go 1.22+                         |
| HTTP       | Fiber v2                         |
| DI         | Uber fx                          |
| Database   | PostgreSQL 16 (pgx/v5 + pgxpool) |
| Migrations | golang-migrate                   |
| Config     | spf13/viper + joho/godotenv      |
| Short code | Snowflake ID → mod 62⁶ → base62  |
| Cache      | Local LRU or Redis (optional)    |
| Logging    | Uber zap (structured JSON)       |

---

## How to Run

### Prerequisites

- Go 1.22+
- Docker & Docker Compose (for the database and Redis)
- `golang-migrate` CLI (for running migrations manually)

```bash
brew install golang-migrate   # macOS
```

### Option A: Docker Compose (recommended)

This spins up the app, Postgres, and Redis together.

```bash
# 1. Copy the example env file
cp .env.example .env

# 2. Start everything
docker compose up --build
```

The service will be available at `http://localhost:8080`.

### Option B: Run locally

```bash
# 1. Start only the infrastructure (Postgres + Redis)
make infra-up

# 2. Copy and configure env
cp .env.example .env
# Edit .env if you need non-default ports or credentials

# 3. Run database migrations
make migrate-up

# 4. Start the server
make run
```

The server reads `config.yaml` first, then overrides from `.env`, then OS env vars.  
Key env var overrides use `__` as the separator (e.g. `SERVER__PORT=9090`).

---

## API Reference

### POST /encode

Shorten a URL.

```bash
curl -X POST http://localhost:8080/encode \
  -H "Content-Type: application/json" \
  -d '{"url": "https://codesubmit.io/library/react"}'
```

**Response 200**

```json
{ "short_url": "http://localhost:8080/GeAi9K" }
```

**Errors**

| Status | Body                                | Reason                          |
| ------ | ----------------------------------- | ------------------------------- |
| 400    | `{"error": "invalid request body"}` | Malformed JSON                  |
| 400    | `{"error": "invalid URL: ..."}`     | Non-HTTP/S scheme, > 2048 chars |
| 500    | `{"error": "internal error"}`       | Unexpected server failure       |

![ENCODE](docs/test-encode.png)

---

### POST /decode

Resolve a short URL back to the original.

```bash
curl -X POST http://localhost:8080/decode \
  -H "Content-Type: application/json" \
  -d '{"short_url": "http://localhost:8080/GeAi9K"}'
```

Both the full short URL and the bare code are accepted:

```json
{ "short_url": "GeAi9K" }
```

**Response 200**

```json
{ "url": "https://codesubmit.io/library/react" }
```

**Errors**

| Status | Body                                | Reason             |
| ------ | ----------------------------------- | ------------------ |
| 400    | `{"error": "invalid request body"}` | Malformed JSON     |
| 404    | `{"error": "short URL not found"}`  | Code doesn't exist |
| 500    | `{"error": "internal error"}`       | Unexpected failure |

![DECODE](docs/test-decode.png)

---

### GET /:code

Browser redirect.

```bash
curl -L http://localhost:8080/GeAi9K
# → 302 Location: https://codesubmit.io/library/react
```

---

## Running Tests

```bash
# Unit + integration tests (no real DB required)
make test

# With race detector
go test -race ./...

# Coverage report (opens HTML in browser)
make test-cover

# Integration tests against a real Postgres instance
DATABASE_URL=postgres://shortlink:shortlink@localhost:5432/shortlink?sslmode=disable \
  go test ./internal/repositories/...
```

### Load tests (k6)

```bash
brew install k6

# Smoke test (default)
make k6-encode
make k6-decode

# Load or stress profile
PROFILE=load BASE_URL=http://localhost:8080 make k6-encode
```

---

## Security Considerations

### 1. URL injection / SSRF

**Risk**: An attacker submits internal URLs (`http://169.254.169.254/...`, `http://localhost/admin`) to probe internal services.  
**Mitigation**: Validation rejects non-HTTP/S schemes. For production, additionally blocklist RFC-1918 address ranges and resolve hostnames before accepting the URL.

### 2. Redirect abuse (open redirector)

**Risk**: Short links can redirect users to phishing or malware sites; the service becomes a reputation-laundering proxy.  
**Mitigation (documented)**: For production, integrate a URL reputation feed (Google Safe Browsing API, VirusTotal) at encode time and on redirect. Add a warning interstitial page for first-time visitors to external domains.

### 3. Enumeration / scraping

**Risk**: 6-character base62 codes can be iterated (~56.8 billion combinations, but random distribution makes exhaustive scraping slow). A brute-force scan could discover every live short link.  
**Mitigation**: Rate-limit both `/encode` and `/decode` per IP (e.g., 60 req/min). Log anomalous decode patterns. For private links, switch from base62 to a longer code or add an optional passphrase.

### 4. Denial of Service

**Risk**: Bulk encode requests flood the database with junk rows; bulk decode requests against non-existent codes hammer Postgres.  
**Mitigation**: Rate limiting (see above). The decode cache absorbs hot paths. Parameterized queries prevent SQL injection from amplifying writes.

### 5. SQL injection

**Risk**: Malicious input in URL or code fields alters queries.  
**Mitigation**: All queries use `pgx` parameterized placeholders (`$1`, `$2`). No string concatenation is used in SQL.

### 6. Logging sensitive data

**Risk**: Long URLs can contain tokens, session cookies, or PII in query parameters; logging them exposes secrets.  
**Mitigation**: The request logger records method, path, status, and latency only — not request bodies or URL values.

### 7. Header / host spoofing

**Risk**: `X-Forwarded-For` or `X-Forwarded-Host` injection if the service sits behind a misconfigured proxy.  
**Mitigation**: Trust proxy headers only from known upstream IPs. Fiber's `TrustedProxies` config should be set explicitly in production.

---

## Scalability & Collision Strategy

### Short code generation

Each code is produced by:

1. Generating a **Snowflake ID** (64-bit, time-ordered, node-scoped — no coordination needed).
2. Taking `id mod 62^6` to project into the ~56.8 billion code space.
3. Base62-encoding the result to a fixed 6-character string.

**Why Snowflake?** IDs are monotonically increasing within a node and across time. This means sequential encodes map to codes that are well-distributed after the modulo step, and no central counter or UUID coordination is needed.

### Collision handling

Because multiple Snowflake IDs can map to the same value after `mod 62^6`, collisions are possible. The service handles them as follows:

1. The DB enforces a `UNIQUE` constraint on `code`.
2. On an `INSERT` conflict, the service retries up to **3 times** with a newly generated Snowflake ID.
3. After 3 failures (astronomically unlikely at typical scale), an error is returned to the caller.

**Collision probability at scale**: At 1 million stored URLs, the birthday-paradox collision probability across 56.8 billion slots is ~0.0009% per insert — negligible. At 100 million URLs (~0.18% of the space filled), a single retry handles virtually every case.

### Horizontal scaling

| Concern                | Current state               | Production path                                                                 |
| ---------------------- | --------------------------- | ------------------------------------------------------------------------------- |
| Multiple app instances | Single node (node ID = 1)   | Assign a unique Snowflake node ID (0–1023) per instance via env var             |
| Write throughput       | Single Postgres primary     | PgBouncer pooling → read replicas for decode → eventually Citus or partitioning |
| Decode hot path        | Optional local LRU or Redis | Redis cluster with short TTL; LRU warms from DB on cache miss                   |
| Migrations             | Run at startup              | Gate behind a leader-election lock or a separate migration job in CI/CD         |
| Rate limiting          | Not yet implemented         | Redis-backed token bucket (e.g., `go-redis/v9` + sliding window) per IP         |

### Code space exhaustion

6 characters of base62 gives **56,800,235,584** unique codes. At 1,000 new URLs per second, the space lasts ~1,800 years. If longer codes are ever needed, increase `codeLen` to 7 (3.5 trillion codes) with a single constant change — no schema migration required since the column is `VARCHAR(6)` padded to the declared length.

---

## My Tasks and AI Assignments

That's a fair that I assign some tasks for AI to help work on this faster and more security.

### My Tasks

- Design structure of project
- Plan step by step the project
- Pick up teck stack for project
- Set up the source from **Config**, **Handler**, **Infra**, **Reposibities**, **Services**, **Utils**, **Logging**, **Database**.
- Implement the core functions such as `/encode` and `/decode` and `/redirect`
- Design Data Model and Migration.
- Plan & apply solutions for the problems such as `thundering-herd` (singleflight) and `cache` and `URL validation`.
- Write *STEPS.md* mannually to share my thoughts.

### AI Assignments

I have a collaboration with AI some aspects in this project

- Generate testcase and all tests.
- Using mockery to mock up repositories and code generator helpers.
- Generate UI and page handler to interact with logic easily.
- Improve Security for UI and for Endpoints like CSP, XSS, SSRF Headers.
- Add config.yaml instead of written by hand.
- Generate Dockerfile and docker compose
- Build Makefile
- Write *README.md* based on *STEPS.md*
- Generate Error Definition and Description

That's some tedious scopes, I think I can hand over AI work for me.
