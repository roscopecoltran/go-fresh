#!/bin/bash

set -eou pipefail
# set -x

# TODO: take name and email for git commit

project="$1"
gitremote="$2"
gitbranch="$3"
dependency="$4"
toversion="$5"
torevision="$6"

projectpath="$GOPATH/src/$project"

mkdir -p "$projectpath"
git clone --depth 1 --branch "$gitbranch" "$gitremote" "$projectpath"
cd "$projectpath"

cat vendor/vendor.json \
    | jq -r '.package[] | select(.path | startswith("'"$dependency"'/")) | .path' \
    | xargs -I {} govendor fetch "{}@${torevision}"

cat vendor/vendor.json \
    | jq -r '.package[] | select(.path | startswith("'"$dependency"'/")) | .path' \
    | xargs -I {} govendor sync "{}"

git add --all

git \
    -c user.name='go-fresh bot' \
    -c user.email='email@example.com' \
    commit --no-verify --no-post-rewrite -m "chore: update $dependency to $toversion"

echo
git log -n 1
