FROM golang:1.18 as builder
WORKDIR /go/src
COPY . .
RUN make test
RUN make build

FROM alpine
RUN mkdir processed_data
COPY --from=builder /go/src/bin/user-aggregator /usr/bin
ENTRYPOINT [ "user-aggregator" ]