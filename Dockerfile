# build stage
FROM golang:1.9.3-alpine3.7

ADD . /go/src/github.com/banzaicloud/spot-termination-collector
WORKDIR /go/src/github.com/banzaicloud/spot-termination-collector
RUN go build -o /bin/spot-termination-collector .

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=0 /bin/spot-termination-collector /bin
ENTRYPOINT ["/bin/spot-termination-collector"]
