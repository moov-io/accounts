FROM golang:1.13-buster as builder
WORKDIR /go/src/github.com/moov-io/accounts
RUN apt-get update && apt-get install make gcc g++
COPY . .
ENV GO111MODULE=on
RUN go mod download
RUN make build

FROM debian:10
RUN apt-get update && apt-get install -y ca-certificates

COPY --from=builder /go/src/github.com/moov-io/accounts/bin/server /bin/server
# USER moov

EXPOSE 8080
EXPOSE 9090
ENTRYPOINT ["/bin/server"]
