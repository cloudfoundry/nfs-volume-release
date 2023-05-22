#########################################################################
# Find, download and uncompress latest version of $STEMCELL_NAME from Bosh.io
#########################################################################
FROM alpine as uncompressed-stemcell
ARG STEMCELL_NAME
RUN apk add curl --no-cache
RUN STEMCELL_URL="$(curl -L https://bosh.io/stemcells | grep -io "https:\/\/.*warden-boshlite-${STEMCELL_NAME}-go_agent.tgz")" \
 && curl -o /stemcell.tgz -L ${STEMCELL_URL}
RUN mkdir -p /wrapper && tar xf /stemcell.tgz -C /wrapper
RUN mkdir -p /stemcell && tar xf /wrapper/image -C /stemcell

#########################################################################
# Run bosh create-release and uncompress tarball which contains blobs and vendored packages
#########################################################################
FROM alpine as uncompressed-release
RUN apk add bash curl git ca-certificates --no-cache
RUN curl -o /bin/bosh -L "https://github.com/cloudfoundry/bosh-cli/releases/download/v7.0.1/bosh-cli-7.0.1-linux-amd64" && chmod +x /bin/bosh

COPY . /release-repo
WORKDIR /release-repo

RUN git submodule update --init --recursive
RUN curl -o /release-old.tgz -L "https://bosh.io/d/github.com/cloudfoundry/nfs-volume-release?v=7.1.6"
RUN bosh create-release --force --tarball /release-new.tgz
RUN mkdir -p /release-old && tar -xf /release-old.tgz -C /release-old
RUN mkdir -p /release-new && tar -xf /release-new.tgz -C /release-new

WORKDIR /release-old/packages
RUN bash -c 'for tgz in *.tgz ; do    \
  filename=$(basename -- "${tgz}") ;  \
  ext="${filename##*.}"            ;  \
  noext="${filename%.*}"           ;  \
                                      \
  mkdir -p "${noext}"              ;  \
  tar -xf "${tgz}" -C "${noext}"   ;  \
  rm "${tgz}"                      ;  \
  echo "${tgz} extracted!"         ;  \
done'

WORKDIR /release-old/jobs
RUN bash -c 'for tgz in *.tgz ; do    \
  filename=$(basename -- "${tgz}") ;  \
  ext="${filename##*.}"            ;  \
  noext="${filename%.*}"           ;  \
                                      \
  mkdir -p "${noext}"              ;  \
  tar -xf "${tgz}" -C "${noext}"   ;  \
  rm "${tgz}"                      ;  \
  echo "${tgz} extracted!"         ;  \
done'

WORKDIR /release-new/packages
RUN bash -c 'for tgz in *.tgz ; do    \
  filename=$(basename -- "${tgz}") ;  \
  ext="${filename##*.}"            ;  \
  noext="${filename%.*}"           ;  \
                                      \
  mkdir -p "${noext}"              ;  \
  tar -xf "${tgz}" -C "${noext}"   ;  \
  rm "${tgz}"                      ;  \
  echo "${tgz} extracted!"         ;  \
done'

WORKDIR /release-new/jobs
RUN bash -c 'for tgz in *.tgz ; do    \
  filename=$(basename -- "${tgz}") ;  \
  ext="${filename##*.}"            ;  \
  noext="${filename%.*}"           ;  \
                                      \
  mkdir -p "${noext}"              ;  \
  tar -xf "${tgz}" -C "${noext}"   ;  \
  rm "${tgz}"                      ;  \
  echo "${tgz} extracted!"         ;  \
done'

#########################################################################
# Render templates
#########################################################################
FROM ruby as rendered-templates
COPY --from=uncompressed-release /release-old /release-old
COPY --from=uncompressed-release /release-new /release-new
WORKDIR /release-new/jobs
RUN mkdir -p /var/vcap/jobs/nfsv3driver-new/bin \
 && gem install bosh-template \
 && bosh-template nfsv3driver/templates/pre-start.erb --context '{"properties":{"nfsv3driver":{"disable":false, "driver_path":"/var/vcap/data/voldrivers"}}}' > /var/vcap/jobs/nfsv3driver-new/bin/pre-start \
 && chmod +x /var/vcap/jobs/nfsv3driver-new/bin/pre-start \
 && mkdir -p /var/vcap/jobs/nfsv3driver/config/certs \
 && touch /var/vcap/jobs/nfsv3driver/config/certs/ca.crt \
 && touch /var/vcap/jobs/nfsv3driver/config/certs/client.crt \
 && touch /var/vcap/jobs/nfsv3driver/config/certs/client.key \
 && touch /var/vcap/jobs/nfsv3driver/config/certs/server.crt \
 && touch /var/vcap/jobs/nfsv3driver/config/certs/server.key
WORKDIR /release-old/jobs
RUN mkdir -p /var/vcap/jobs/nfsv3driver-old/bin \
 && gem install bosh-template \
 && bosh-template nfsv3driver/templates/install.erb --context '{"properties":{"nfsv3driver":{"disable":false, "driver_path":"/var/vcap/data/voldrivers"}}}' > /var/vcap/jobs/nfsv3driver-old/bin/pre-start \
 && chmod +x /var/vcap/jobs/nfsv3driver-old/bin/pre-start \
 && mkdir -p /var/vcap/jobs/nfsv3driver-old/config/certs \
 && touch /var/vcap/jobs/nfsv3driver-old/config/certs/ca.crt \
 && touch /var/vcap/jobs/nfsv3driver-old/config/certs/client.crt \
 && touch /var/vcap/jobs/nfsv3driver-old/config/certs/client.key \
 && touch /var/vcap/jobs/nfsv3driver-old/config/certs/server.crt \
 && touch /var/vcap/jobs/nfsv3driver-old/config/certs/server.key

#########################################################################
# Mount alongside stemcell and uncompressed release files and run packaging scripts
#########################################################################
FROM scratch as compiled-release
COPY --from=uncompressed-stemcell /stemcell/ /
COPY --from=uncompressed-release /release-old /release-old
COPY --from=uncompressed-release /release-new /release-new
ENV PACKAGES="golang-1-linux          \
nfs-debs                              \
nfsbroker                             \
nfsv3driver                           \
berkeleydb                            \
openldap-2.6.3"


WORKDIR /release-old/packages
RUN bash -c 'for pkg in ${PACKAGES}; do                   \
 export BOSH_INSTALL_TARGET="/var/vcap/packages/${pkg}" ; \
 export BOSH_COMPILE_TARGET="/release-old/packages/${pkg}"  ; \
 cd "/release-old/packages/${pkg}"                          ; \
 mkdir -p "${BOSH_INSTALL_TARGET}"                      ; \
 chmod +x packaging                                     ; \
 ./packaging                                            ; \
done'
ENV PACKAGES="golang-1-linux          \
nfs-debs                              \
nfsbroker                             \
nfsv3driver                           \
berkeleydb                            \
openldap-2.5.13"


WORKDIR /release-new/packages
RUN bash -c 'for pkg in ${PACKAGES}; do                   \
 export BOSH_INSTALL_TARGET="/var/vcap/packages/${pkg}" ; \
 export BOSH_COMPILE_TARGET="/release-new/packages/${pkg}"  ; \
 cd "/release-new/packages/${pkg}"                          ; \
 mkdir -p "${BOSH_INSTALL_TARGET}"                      ; \
 chmod +x packaging                                     ; \
 ./packaging                                            ; \
done'


#########################################################################
# Ready to run, clean, filesystem with the stemcell and compiled packages in place
#########################################################################
FROM scratch
COPY --from=uncompressed-stemcell /stemcell/ /
COPY --from=rendered-templates /var/vcap/ /var/vcap/
COPY --from=compiled-release /var/vcap/ /var/vcap/
RUN chown -R man: /var/cache/man/ \
 && chmod -R 755 /var/cache/man/ 

ENTRYPOINT ["/bin/bash", "-c", "/var/vcap/jobs/nfsv3driver-old/bin/pre-start && /var/vcap/jobs/nfsv3driver-new/bin/pre-start"]

