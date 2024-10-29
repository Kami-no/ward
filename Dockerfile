# build stage
FROM golang:1.23-alpine3.15 AS build-env
# hadolint ignore=DL3018
RUN apk add --no-cache git

WORKDIR /opt

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .
# RUN go test ./...
# hadolint ignore=DL3059
RUN GOPROXY=http://172.17.0.1:8080,direct GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOFLAGS=-insecure go build -o ward

# final stage
# hadolint ignore=DL3007
FROM alpine:3.15
WORKDIR /app
COPY --from=build-env /opt/templates /app/
COPY --from=build-env /opt/ward /app/
CMD ["./ward"]
