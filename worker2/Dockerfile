# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /src
# Copy the entire project
COPY . .
# Build only worker2/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker2 ./worker2/main.go

# Final stage
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/worker2 /app/worker2
EXPOSE 8001 2113
ENTRYPOINT ["/app/worker2"]
