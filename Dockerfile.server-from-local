# Build stage
ARG GO_VERSION=1.10
ARG PROJECT_PATH=/go/src/github.com/dustin-decker/threatseer
FROM golang:${GO_VERSION}-alpine AS builder
RUN apk --no-cache add ca-certificates
RUN adduser -D -u 59999 container-user
WORKDIR /go/src/github.com/dustin-decker/threatseer
COPY ./ ${PROJECT_PATH}

# Production image
FROM scratch
EXPOSE 8081
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/dustin-decker/threatseer/threatseer.yml /threatseer.yml
COPY --from=builder /go/src/github.com/dustin-decker/threatseer/config /config
COPY --from=builder /go/src/github.com/dustin-decker/threatseer/bin/server /bin/server
ENTRYPOINT ["/bin/server"]
