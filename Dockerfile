FROM luisbebop/go1.2

ENV PATH $PATH:$GOPATH/bin:$GOROOT/bin

ADD server /opt/server

EXPOSE 8800

CMD ["/opt/server"]
