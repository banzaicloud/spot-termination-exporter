# build stage
FROM golang:1.9.3-alpine3.7

ADD . /go/src/github.com/banzaicloud/spot-termination-exporter
WORKDIR /go/src/github.com/banzaicloud/spot-termination-exporter
RUN go build -o /bin/spot-termination-exporter .

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=0 /bin/spot-termination-exporter /bin

USER nobody

ENTRYPOINT ["/bin/spot-termination-exporter"]
