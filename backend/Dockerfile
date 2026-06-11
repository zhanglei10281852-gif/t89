FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o wedding-server .

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/wedding-server /app/

VOLUME ["/app/data"]

EXPOSE 8627

CMD ["/app/wedding-server"]
