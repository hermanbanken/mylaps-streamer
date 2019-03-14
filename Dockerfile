ARG BASE=alpine
FROM golang:1.11.2-alpine AS builder
RUN apk add --no-cache git
WORKDIR /src
ADD main.go go.mod go.sum /src/
RUN go build -o /main

FROM $BASE
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /main /

CMD /main