# Build stage
ARG GO_VERSION=1.10
ARG PROJECT_PATH=/go/src/github.com/dustin-decker/threatseer
FROM golang:${GO_VERSION}-alpine AS builder
RUN apk --no-cache add git ca-certificates
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep
RUN adduser -D -u 59999 container-user
RUN apk --no-cache add gcc libc-dev 
WORKDIR /go/src/github.com/dustin-decker/threatseer
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY ./ ${PROJECT_PATH}
RUN export PATH=$PATH:`go env GOHOSTOS`-`go env GOHOSTARCH` \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build --ldflags '-extldflags "-static"' -o bin/agent agent/*.go \
    && go test $(go list ./... | grep -v /vendor/)

# Production image
FROM scratch
EXPOSE 8081
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/dustin-decker/threatseer/bin/agent /bin/agent
ENTRYPOINT ["/bin/agent"]