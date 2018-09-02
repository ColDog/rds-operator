#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

vendor/k8s.io/code-generator/generate-groups.sh \
deepcopy \
github.com/coldog/rds-operator/pkg/generated \
github.com/coldog/rds-operator/pkg/apis \
rds:v1alpha1 \
--go-header-file "./build/codegen/boilerplate.go.txt"
