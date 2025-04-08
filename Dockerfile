# Stage 1: Build
FROM golang:1.21.13-bookworm AS builder

WORKDIR /app

# Copy go.mod dan go.sum lebih awal (agar caching optimal)
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh isi project
COPY . .

# Build binary dari folder cmd/
WORKDIR /app/cmd
RUN go build -o main .

# Stage 2: Runtime
FROM debian:bookworm-slim

WORKDIR /root/

# Copy binary dari stage sebelumnya
COPY --from=builder /app/cmd/main .

# Install ca-certificates & git
RUN apt-get update && apt-get install -y ca-certificates git

# Copy CA certs ke image final
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 9000

CMD ["./main"]
