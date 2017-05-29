FROM alpine:3.6

ARG HELM_VERSION=2.4.2

RUN apk --no-cache add \
    ca-certificates \
    wget \
    py2-pip

# install aws cli
RUN pip install awscli

# install helm
RUN \
    mkdir -p /opt/helm && \
    wget -O /opt/helm/helm.tar.gz https://kubernetes-helm.storage.googleapis.com/helm-v${HELM_VERSION}-linux-amd64.tar.gz && \
    tar zxvf /opt/helm/helm.tar.gz -C /opt/helm && \
    mv /opt/helm/linux-amd64/* /opt/helm/ && \
    ln -s /opt/helm/helm /usr/local/bin/helm && \
    rm -rf /opt/helm/helm.tar.gz && rmdir /opt/helm/linux-amd64

COPY build/hrp /opt/hrp

ENTRYPOINT ["/opt/hrp"]