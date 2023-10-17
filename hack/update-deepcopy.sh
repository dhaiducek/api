#!/bin/bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}

verify="${VERIFY:-}"

# cluster:v1alpha1 is handed separately with controller-gen in the Makefile
# because GenGo doesn't currently support generics: https://github.com/kubernetes/gengo/issues/225
GOFLAGS="" bash ${CODEGEN_PKG}/generate-groups.sh "deepcopy" \
  open-cluster-management.io/api/generated \
  open-cluster-management.io/api \
  "cluster:v1,v1beta1,v1beta2 work:v1alpha1,v1 operator:v1 addon:v1alpha1" \
  --go-header-file ${SCRIPT_ROOT}/hack/empty.txt \
  ${verify}
