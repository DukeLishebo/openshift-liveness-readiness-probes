FROM golang:1.16-alpine

RUN apk add --no-cache curl

WORKDIR /go/src/app
COPY go.mod .
COPY main.go .
RUN go build -o app

ENTRYPOINT ./app
