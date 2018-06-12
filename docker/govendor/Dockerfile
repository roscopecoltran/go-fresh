FROM golang:1.10-alpine3.7

RUN apk add --no-cache \
    git \
    bash \
    jq \
    openssh \
    curl

RUN set -x \
    && go get github.com/kardianos/govendor \
    \
    && git version \
    && go version \
    && govendor --version

COPY askpass.sh pr.sh /opt/

# TODO: add a stepdown?
ENTRYPOINT ["/opt/pr.sh"]
