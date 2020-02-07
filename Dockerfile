# build stage
FROM golang:alpine AS build-env
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ward

# final stage
# hadolint ignore=DL3007
FROM alpine:latest
WORKDIR /app
COPY --from=build-env /go/ward /app/
CMD ["./ward"]
