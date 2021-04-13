# Build Geth in a stock Go builder container
FROM golang:1.10-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /aqchain
RUN cd /aqchain && make aqchain

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /aqchain/build/bin/aqchain /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["aqchain"]
