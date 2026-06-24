# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.23-bookworm AS build
WORKDIR /src

# Constrain Go's parallelism/memory so the build doesn't get OOM-killed
# (exit 137) on memory-limited build runners.
ENV GOMAXPROCS=2 \
    GOMEMLIMIT=384MiB

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/docflow-backend .

# ---- runtime stage ----
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        libreoffice-core \
        libreoffice-writer \
        libreoffice-calc \
        libreoffice-impress \
        fonts-liberation \
        fonts-dejavu \
        ca-certificates \
        python3 \
        python3-pip \
        libgl1 \
        libglib2.0-0 \
    && pip3 install --no-cache-dir --break-system-packages \
        pdf2docx "PyMuPDF<1.24" python-pptx pdfplumber openpyxl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/docflow-backend /usr/local/bin/docflow-backend

ENV PORT=8080 \
    SOFFICE_BIN=soffice \
    PYTHON_BIN=python3 \
    WORK_DIR=/tmp/docflow

RUN useradd --create-home --uid 10001 app \
    && mkdir -p /tmp/docflow \
    && chown -R app /tmp/docflow
USER app

EXPOSE 8080
ENTRYPOINT ["docflow-backend"]
