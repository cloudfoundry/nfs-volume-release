FROM relintdockerhubpushbot/cf-deployment-concourse-tasks as golang_version
RUN git clone --recurse-submodules https://github.com/cloudfoundry/nfs-volume-release
RUN cd nfs-volume-release && bosh create-release --tarball /tmp/release.tgz
RUN version=$(cat /tmp/release.tgz | tar -Oxz packages/golang-1.13-linux.tgz | tar z --list | grep -ohE "go[0-9]\.[0-9]{1,2}\.[0-9]{1,2}") && echo $version > /tmp/golang_version

FROM openjdk:8 as builder
RUN git clone https://github.com/cloudfoundry-incubator/credhub
RUN cd credhub && ./gradlew downloadBouncyCastleFips assemble

FROM openjdk:8
RUN apt-get update
RUN apt-get install -y gcc
COPY --from=golang_version /tmp/golang_version /tmp/golang_version
RUN wget https://dl.google.com/go/$(cat /tmp/golang_version).linux-amd64.tar.gz
RUN tar -xvf $(cat /tmp/golang_version).linux-amd64.tar.gz
RUN mv go /usr/local
ENV GOROOT=/usr/local/go
ENV PATH=$PATH:/usr/local/go/bin:~/go/bin

RUN go get github.com/onsi/ginkgo/ginkgo
COPY --from=builder /root/.gradle/ /root/.gradle/