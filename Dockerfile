FROM golang:1.20-bullseye

COPY ./src /opt/src
WORKDIR /opt/src

RUN apt-get update \
    && apt-get install -y libpcap-dev \
    && go mod download \
    && go build -trimpath -ldflags "-w -s" -o /build/app

CMD ["/build/app"]
