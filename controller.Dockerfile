FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY controller-binary /app/controller
CMD ["/app/controller"]