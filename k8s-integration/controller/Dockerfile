FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN apk add --no-cache git
ENV GOPROXY=direct
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/controller ./k8s-integration/controller/main.go 

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /bin/controller /app/controller
ENTRYPOINT ["/app/controller"]