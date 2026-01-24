# ---------- Build stage ----------

FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (needed for go modules sometimes)
RUN apk add --no-cache git

# Copy go mod files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Tidy and build the binary
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/api


# ---------- Runtime stage ----------
FROM alpine:3.19

WORKDIR /app

# Certificates for HTTPS / TLS
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/app .

# Expose app port
EXPOSE 8080

# Run
CMD ["./app"]
