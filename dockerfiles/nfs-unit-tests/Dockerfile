FROM relintdockerhubpushbot/cf-deployment-concourse-tasks as golang_version
RUN git clone --recurse-submodules https://github.com/cloudfoundry/nfs-volume-release
RUN cd nfs-volume-release && bosh create-release --tarball /tmp/release.tgz
RUN version=$(cat /tmp/release.tgz | tar -Oxz packages/golang-1.13-linux.tgz | tar z --list | grep -ohE "go[0-9]\.[0-9]{1,2}\.[0-9]{1,2}") && echo $version > /tmp/golang_version
RUN ginkgo_version=$(cd nfs-volume-release/src/bosh_release/ && go list -f "{{.Version}}" -m github.com/onsi/ginkgo > /tmp/ginkgo_version)

FROM ubuntu

RUN apt-get update && apt-get install -y git gcc curl wget

COPY --from=golang_version /tmp/golang_version /tmp/golang_version
COPY --from=golang_version /tmp/ginkgo_version /tmp/ginkgo_version
RUN wget https://dl.google.com/go/$(cat /tmp/golang_version).linux-amd64.tar.gz
RUN tar -xvf $(cat /tmp/golang_version).linux-amd64.tar.gz
RUN mv go /usr/local

ENV PATH "$PATH:/usr/local/go/bin"
ENV GOPATH /go
ENV GOROOT=/usr/local/go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
RUN GO111MODULE=on go get github.com/onsi/ginkgo/ginkgo@$(cat /tmp/ginkgo_version)
