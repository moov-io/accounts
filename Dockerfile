FROM golang:1.15-buster as builder
WORKDIR /go/src/github.com/moov-io/accounts
RUN apt-get update && apt-get install make gcc g++
COPY . .
RUN go mod download
RUN make build

FROM debian:10
MAINTAINER Moov <support@moov.io>

RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /go/src/github.com/moov-io/accounts/bin/server /bin/server

# USER moov
EXPOSE 8080
EXPOSE 9090
ENTRYPOINT ["/bin/server"]
