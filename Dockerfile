# Stage 1: Build the Go binary
FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o main .


FROM alpine:latest

RUN apk --no-cache add ca-certificates curl tzdata

ENV TZ=UTC

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

RUN mkdir -p /app/storage/uploads /app/storage/processed && \
    chown -R appuser:appgroup /app/storage

COPY --from=builder /app/main .

RUN chown appuser:appgroup main

USER appuser

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

CMD ["./main"]