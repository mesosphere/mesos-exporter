FROM golang:alpine3.7

WORKDIR /go/src/github.com/mesosphere/mesos_exporter

COPY . .

RUN go build -o /bin/mesos-exporter


FROM alpine:3.7

COPY --from=0 /bin/mesos-exporter /bin/mesos-exporter

EXPOSE 9105

RUN addgroup exporter &&\
    adduser -S -G exporter exporter &&\
    apk --update add ca-certificates &&\
    rm -rf /var/cache/apk/*

USER exporter

ENV MODE ${MODE}
ENV URL ${URL}

ENTRYPOINT ./bin/mesos-exporter -master "$URL"
