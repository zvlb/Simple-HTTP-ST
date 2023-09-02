##
## Build
##
FROM golang:1.20-alpine3.18 AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -o bin/simple-http-st ./main.go

##
## Deploy
##
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/bin/simple-http-st .

ENTRYPOINT ["/app/simple-http-st"]