FROM ubuntu:16.04

ADD server /opt/server

EXPOSE 8800

CMD ["/opt/server"]
