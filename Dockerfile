# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.23-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/docflow-backend .

# ---- runtime stage ----
FROM debian:bookworm-slim

# LibreOffice (headless) + fonts for faithful rendering.
RUN apt-get update && apt-get install -y --no-install-recommends \
        libreoffice-core \
        libreoffice-writer \
        libreoffice-calc \
        libreoffice-impress \
        fonts-liberation \
        fonts-dejavu \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/docflow-backend /usr/local/bin/docflow-backend

ENV PORT=8080 \
    SOFFICE_BIN=soffice \
    WORK_DIR=/tmp/docflow

RUN useradd --create-home --uid 10001 app \
    && mkdir -p /tmp/docflow \
    && chown -R app /tmp/docflow
USER app

EXPOSE 8080
ENTRYPOINT ["docflow-backend"]
