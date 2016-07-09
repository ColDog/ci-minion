FROM golang:latest

RUN mkdir /app

RUN go version
ADD . /go/src/app/
WORKDIR /go/src/app/

RUN echo $GOPATH
RUN go get ./...
RUN go build -o main .
CMD ["/app/main"]
