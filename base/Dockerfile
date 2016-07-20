FROM ubuntu:latest

ADD . /install

WORKDIR /install/

RUN install/basics.sh
RUN install/docker.sh
RUN install/git.sh
RUN install/java.sh
RUN install/node.sh
RUN install/python.sh
RUN install/ruby.sh

RUN mkdir -p /opt/ci
WORKDIR /opt/ci

ENTRYPOINT ["bash"]
