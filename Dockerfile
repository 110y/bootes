FROM golang:1.14.0-alpine3.11 AS builder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /go/src/github.com/110y/bootes

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN go build -o /usr/bin/bootes .

# runtime image
FROM alpine:3.11.3

COPY --from=builder /usr/bin/bootes /usr/bin/bootes
RUN apk update
RUN apk add --no-cache ca-certificates

EXPOSE 8080

CMD ["/usr/bin/bootes"]
