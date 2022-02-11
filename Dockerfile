ARG GO_VERSION=1.16

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*
WORKDIR /checker

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY checker/ ./
RUN go build -o ./app ./main.go

FROM alpine:latest

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /checker
COPY --from=builder /checker/app .

ENTRYPOINT ["./app"]