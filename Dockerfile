# build stage
FROM golang:alpine AS build-env
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gl-watcher

# final stage
# hadolint ignore=DL3007
FROM alpine:latest
WORKDIR /app
COPY --from=build-env /go/gl-watcher /app/
CMD ["./gl-watcher"]
