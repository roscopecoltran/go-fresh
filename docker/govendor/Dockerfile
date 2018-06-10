FROM golang:1.10-alpine3.7

RUN apk add --no-cache \
    git \
    bash \
    jq

# TODO: add a stepdown?

RUN set -x \
    && go get github.com/kardianos/govendor \

    && git version \
    && go version \
    && govendor --version

COPY pr.sh /opt/

ENTRYPOINT ["/opt/pr.sh"]

# build:
# docker build -t go-fresh-govendor .

# test:
# docker run --rm go-fresh-govendor \
#   "github.com/terraform-providers/terraform-provider-azurerm" \
#   "https://github.com/terraform-providers/terraform-provider-azurerm.git" \
#   "master" \
#   "golang.org/x/text" \
#   "0.3.0" \
#   "f21a4dfb5e38f5895301dc265a8def02365cc3d0"