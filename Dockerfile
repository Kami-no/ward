# build stage
FROM golang:1.15-alpine AS build-env
WORKDIR /opt
COPY . .
RUN go test ./...
RUN GOPROXY=http://172.17.0.1:8080,direct GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOFLAGS=-insecure go build -o ward

# final stage
# hadolint ignore=DL3007
FROM alpine:latest
WORKDIR /app
COPY --from=build-env /opt/templates /app/
COPY --from=build-env /opt/config.yaml /app/
COPY --from=build-env /opt/ward /app/
CMD ["./ward"]
