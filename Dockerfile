# Stage 1: Build
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go build -ldflags="-w -s" -o server ./cmd/app

# Stage 2: Run
FROM alpine:3.22

RUN adduser -D appuser
WORKDIR /home/appuser
COPY --from=builder --chown=appuser /app/server .
USER appuser

EXPOSE 8080
CMD ["./server"]