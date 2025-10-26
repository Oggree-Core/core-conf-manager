FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the Go source file
COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o configManager main.go


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/configManager .

ENTRYPOINT ["/app/configManager"]
