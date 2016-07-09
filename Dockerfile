FROM ubuntu

RUN apt-get update
RUN apt-get install -y git

RUN mkdir -p /opt/ci
WORKDIR /opt/ci
