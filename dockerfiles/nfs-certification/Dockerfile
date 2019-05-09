FROM janeczku/nfs-ganesha

RUN yum --disablerepo=centos-gluster37 install -y fuse

RUN curl -L https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz | tar -C /usr/local -xz

ENV PATH "$PATH:/usr/local/go/bin"
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN go get github.com/onsi/ginkgo/ginkgo