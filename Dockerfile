# syntax=docker/dockerfile:1

FROM golang:1.26-alpine

RUN apk add --no-cache ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /usr/local/bin/action ./cmd/action

ENTRYPOINT ["/usr/local/bin/action"]
