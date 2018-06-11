#!/bin/bash

set -eou pipefail

# TODO: take name and email for git commit

project="$1"
gitremote="$2"
gitbranch="$3"
dependency="$4"
toversion="$5"
torevision="$6"

projectpath="$GOPATH/src/$project"
prbranch="go-fresh/${dependency}/${toversion}"
commitmsg="chore: update $dependency to $toversion"
prtitle="[go-fresh] Update $dependency to $toversion"

mkdir -p "$projectpath"
git clone --depth 1 --branch "$gitbranch" "$gitremote" "$projectpath"
cd "$projectpath"

echo
echo "Running govendor to update $dependency to $toversion"

cat vendor/vendor.json \
    | jq -r '.package[] | select(.path | startswith("'"$dependency"'/")) | .path' \
    | xargs -I {} govendor fetch "{}@${torevision}"

cat vendor/vendor.json \
    | jq -r '.package[] | select(.path | startswith("'"$dependency"'/")) | .path' \
    | xargs -I {} govendor sync "{}"

echo
echo "Creating commit"

git add --all

git \
    -c user.name="$GIT_USER_NAME" \
    -c user.email="$GIT_USER_EMAIL" \
    commit --no-verify --no-post-rewrite -m "$commitmsg"

echo
git log -n 1

echo
echo "Pushing to $gitremote"

git \
    -c "credential.https://github.com.username"="$GITHUB_USERNAME" \
    -c core.askPass="/opt/askpass.sh" \
    push "$gitremote" "$(git rev-parse --abbrev-ref HEAD):$prbranch"

# TODO: do a real wait for the branch
# eventual consistency!!!
sleep 10

echo
echo "Creating PR"

github_repo="$(echo "$gitremote" | awk -F/ '{print $4 "/" $5}')"
github_repo="${github_repo%.git}"

curl -s -XPOST "https://api.github.com/repos/${github_repo}/pulls" \
-u "$GITHUB_USERNAME:$GITHUB_TOKEN" \
-H 'Accept: application/vnd.github.v3+json' \
-H 'Content-Type: application/json; charset=utf-8' \
-d @- <<JSON | jq .
{
    "title": "$prtitle",
    "head":  "$prbranch",
    "base":  "$gitbranch",

    "maintainer_can_modify": true,
    
    "body": "body..."
}
JSON
