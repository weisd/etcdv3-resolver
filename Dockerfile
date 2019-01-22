FROM golang:alpine

# RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# RUN adduser -D -g '' appuser



COPY ./bin /go/bin



