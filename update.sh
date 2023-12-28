#!/bin/env bash

set -euxo pipefail

version=$1
version_dir=cca-$version
doc_dir=$(go list -f '{{ .Dir }}' -m bitbucket.org/infomaker/doc-format/v2)

if [[ ! -d $version_dir ]]; then
    git clone --depth 1 --branch $version git@bitbucket.org:infomaker/cca.git $version_dir
fi

clean_up () {
    rm -fr $version_dir
}
trap clean_up EXIT

cp $doc_dir/rpc/document.proto ./rpc/document.proto
cp cca-$version/rpc/cca/service.proto ./rpc/cca/service.proto
cp cca-$version/rpc/cca/service.*.go ./

doc_package=$(cd $version_dir && go mod graph | awk '$1=="bitbucket.org/infomaker/cca" {print $2}' | awk -F'@' '$1=="bitbucket.org/infomaker/doc-format/v2"')
go get $doc_package
go mod tidy -go=1.17

git add -u

git commit -m "updated to cca ${version}"
git tag $version
