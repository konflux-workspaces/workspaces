FROM registry.access.redhat.com/ubi8/ubi:latest
ENV LANG=en_US.utf8 \
    GOPATH=/tmp/go \
    GOCACHE=/tmp/.cache \
    PATH=$PATH:$GOPATH/bin \
    GOLANG_VERSION=go1.22.2 \
    GOLANG_SHA256=5901c52b7a78002aeff14a21f93e0f064f74ce1360fce51c6ee68cd471216a17

RUN yum install -y --nodocs \
    gcc \
    git \
    jq \
    make \
    tar \
    skopeo \
    && yum clean all
RUN ln -s /cli/oc /usr/bin/kubectl

WORKDIR /tmp

# download, verify and install golang
ENV PATH=$PATH:/usr/local/go/bin
RUN curl -Lo ${GOLANG_VERSION}.linux-amd64.tar.gz https://dl.google.com/go/${GOLANG_VERSION}.linux-amd64.tar.gz \
    && echo "${GOLANG_SHA256} ${GOLANG_VERSION}.linux-amd64.tar.gz" > ${GOLANG_VERSION}.linux-amd64.sha256 \
    && sha256sum -c ${GOLANG_VERSION}.linux-amd64.sha256 \
    && tar xzf ${GOLANG_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf ${GOLANG_VERSION}.linux-amd64.tar.gz \
    && rm -f ${GOLANG_VERSION}.linux-amd64.tar.gz \
    && go version

ENTRYPOINT [ "/bin/bash" ]
