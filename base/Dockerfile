FROM ubuntu:latest

RUN mkdir -p /install
WORKDIR /install

COPY install/basics.sh /install/basics.sh
RUN ./basics.sh

COPY install/docker.sh /install/docker.sh
RUN ./docker.sh

COPY install/git.sh /install/git.sh
RUN ./git.sh

COPY install/java.sh /install/java.sh
RUN ./java.sh

COPY install/node.sh /install/node.sh
RUN ./node.sh

COPY install/python.sh /install/python.sh
RUN ./python.sh

COPY install/ruby.sh /install/ruby.sh
RUN ./ruby.sh

RUN mkdir -p /opt/ci
WORKDIR /opt/ci

ADD simpleci /usr/local/bin/

RUN rm /bin/sh && ln -s /bin/bash /bin/sh && sed -i 's/^mesg n$/tty -s \&\& mesg n/g' /root/.profile

EXPOSE 8000
CMD ["simpleci", "server"]
