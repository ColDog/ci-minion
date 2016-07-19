FROM golang:latest

RUN apt-get update -qq && apt-get install -qqy \
    apt-transport-https \
    ca-certificates \
    curl \
    lxc \
    iptables

RUN curl -sSL https://get.docker.com/ | sh
#RUN sudo usermod -aG docker root

ADD . /go/src/app/
WORKDIR /go/src/app/

RUN go get ./...
RUN go build -o /go/bin/main .

RUN go version

EXPOSE 8000
CMD ["/go/bin/main"]
