# DocFlow Backend

Self-hosted document conversion microservice for the **DocFlow** mobile app.

Office ↔ PDF conversions run through a headless **LibreOffice** process; PDF page
operations (merge / split / compress) use the pure-Go
[pdfcpu](https://github.com/pdfcpu/pdfcpu) library. The service is stateless —
every request gets an isolated temp workspace that is deleted when the response
is sent.

This is a standalone repository, separate from the Flutter app. The app's
`CloudConversionService` calls these endpoints over HTTP.

---

## API

All conversion endpoints require an `X-API-Key` header matching the `API_KEY`
env var (unless `API_KEY` is empty, which disables auth for local dev).

| Method | Path                    | Body (multipart/form-data)              | Response                |
| ------ | ----------------------- | --------------------------------------- | ----------------------- |
| GET    | `/health`               | —                                       | `{"status":"ok"}`       |
| POST   | `/api/v1/convert`       | `file`, `target` (pdf/docx/xlsx/pptx/…) | converted file (binary) |
| POST   | `/api/v1/pdf/merge`     | `files` (≥2 PDFs)                       | merged PDF              |
| POST   | `/api/v1/pdf/split`     | `file` (1 PDF)                          | zip of single pages     |
| POST   | `/api/v1/pdf/compress`  | `file` (1 PDF)                          | optimized PDF           |

Errors return JSON: `{"error":"..."}` with an appropriate status code
(`400` bad request, `401` unauthorized, `422` conversion failed).

### Supported `target` formats

`pdf`, `docx`, `xlsx`, `pptx`, `txt`, `odt`, `ods`, `odp`, `rtf`, `csv`, `html`.

> **Note on PDF → Office:** LibreOffice converts PDF *into* editable Office
> formats by importing the PDF into Draw first. Results are reliable for
> Office → PDF, but PDF → Word/Excel/PowerPoint is best-effort — complex layouts
> may not round-trip cleanly. This is a LibreOffice limitation, not a bug here.

---

## Run locally

Requires Go 1.23+ and LibreOffice installed (`soffice` on PATH).

```bash
cp .env.example .env        # then edit API_KEY
make tidy                   # resolve dependencies
make run                    # starts on :8080
```

Without local LibreOffice, just use Docker (below) — it bundles everything.

## Run with Docker

```bash
export API_KEY=$(openssl rand -hex 24)
make docker-up              # build + start in background
# ...
make docker-down
```

---

## Examples

```bash
# Word -> PDF
curl -sS -X POST http://localhost:8080/api/v1/convert \
  -H "X-API-Key: $API_KEY" \
  -F "file=@report.docx" \
  -F "target=pdf" \
  -o report.pdf

# Merge PDFs
curl -sS -X POST http://localhost:8080/api/v1/pdf/merge \
  -H "X-API-Key: $API_KEY" \
  -F "files=@a.pdf" -F "files=@b.pdf" \
  -o merged.pdf

# Split into single pages (zip)
curl -sS -X POST http://localhost:8080/api/v1/pdf/split \
  -H "X-API-Key: $API_KEY" \
  -F "file=@book.pdf" \
  -o pages.zip

# Compress
curl -sS -X POST http://localhost:8080/api/v1/pdf/compress \
  -H "X-API-Key: $API_KEY" \
  -F "file=@big.pdf" \
  -o small.pdf
```

---

## Configuration

| Env var               | Default        | Description                              |
| --------------------- | -------------- | ---------------------------------------- |
| `PORT`                | `8080`         | HTTP listen port                         |
| `API_KEY`             | _(empty)_      | Required `X-API-Key`; empty = auth off   |
| `SOFFICE_BIN`         | `soffice`      | Path to LibreOffice binary               |
| `MAX_UPLOAD_MB`       | `50`           | Max request body size                    |
| `CONVERT_TIMEOUT_SEC` | `120`          | Per-conversion timeout                   |
| `WORK_DIR`            | `$TMPDIR/docflow` | Temp workspace root                   |

---

## Deploy

The image is self-contained. Any host with Docker works (VPS, Fly.io, Render,
Cloud Run, etc.). Put it behind HTTPS (reverse proxy / platform TLS) and set a
strong `API_KEY`. Then point the Flutter app's cloud base URL at it.

## Project layout

```
.
├── main.go                      # bootstrap, graceful shutdown, -healthcheck
├── internal/
│   ├── config/                  # env-based configuration
│   ├── converter/               # LibreOffice + pdfcpu logic
│   ├── handler/                 # HTTP handlers + multipart/streaming helpers
│   └── server/                  # routing + middleware (auth, logging, recover)
├── Dockerfile
├── docker-compose.yml
└── Makefile
```
