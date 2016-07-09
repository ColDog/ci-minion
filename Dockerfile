FROM golang:latest

RUN mkdir /app

ADD . /app/
WORKDIR /app

RUN echo $GOPATH
RUN go get ./...
RUN go build -o main .
CMD ["/app/main"]
ex