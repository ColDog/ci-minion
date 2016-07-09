FROM golang:1.6

RUN apt-get update
RUN apt-get -y install apt-transport-https ca-certificates
RUN apt-get -y install docker.io

ADD . /go/src/github.com/golang/coldog/minion

RUN go install github.com/golang/coldog/minion

ENTRYPOINT /go/bin/minion

EXPOSE 8000
