ARG ARCH="amd64"
ARG OS="linux"

FROM golang as builder

WORKDIR /build
COPY . /build/
ENV CGO_ENABLED=0
RUN go install
RUN go build

FROM alpine
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
COPY --from=builder /build/snmp_exporter  /bin/snmp_exporter
COPY snmp.yml       /etc/snmp_exporter/snmp.yml

EXPOSE      9116
ENTRYPOINT  [ "/bin/snmp_exporter" ]
CMD         [ "--config.file=/etc/snmp_exporter/snmp.yml" ]
