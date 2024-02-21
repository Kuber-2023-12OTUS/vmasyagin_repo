#!/bin/sh

namespace=$1
sa=$2
server=${3:-https://localhost:32769}

secret=$(kubectl -n $namespace get secret | grep $sa-token | awk '{print $1}')

ca=$(kubectl -n $namespace get secret/$secret -o jsonpath='{.data.ca\.crt}')
token=$(kubectl -n $namespace get secret/$secret -o jsonpath='{.data.token}' | base64 --decode)
namespace=$(kubectl -n $namespace get secret/$secret -o jsonpath='{.data.namespace}' | base64 --decode)

echo "
apiVersion: v1
kind: Config
clusters:
- name: default-cluster
  cluster:
    certificate-authority-data: ${ca}
    server: ${server}
contexts:
- name: default-context
  context:
    cluster: default-cluster
    namespace: default
    user: ${sa}-user
current-context: default-context
users:
- name: ${sa}-user
  user:
    token: ${token}
" > kubeconfig