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

# build:
# docker build -t go-fresh-govendor .

# test:
# docker run --rm \
#   -e 'GIT_USER_NAME=go-fresh bot' \
#   -e 'GIT_USER_EMAIL=email@example.com' \
#   -e 'GITHUB_USERNAME=go-fresh-dev' \
#   -e "GITHUB_TOKEN=$GITHUB_TOKEN" \
#   go-fresh-govendor \
#   "github.com/terraform-providers/terraform-provider-azurerm" \
#   "https://github.com/terraform-providers/terraform-provider-azurerm.git" \
#   "master" \
#   "golang.org/x/text" \
#   "0.3.0" \
#   "f21a4dfb5e38f5895301dc265a8def02365cc3d0"