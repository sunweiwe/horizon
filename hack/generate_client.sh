#!/bin/bash

set -e

GV="$1"

rm -rf ./pkg/client
./hack/generate_group.sh "client,lister,informer" github.com/sunweiwe/horizon/pkg/client github.com/sunweiwe/api "${GV}" --output-base=./  -h "$PWD/hack/boilerplate.go.txt"
mv  github.com/sunweiwe/horizon/pkg/client ./pkg/
rm -rf ./github.com