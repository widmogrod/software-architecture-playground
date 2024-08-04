# Flink on k8s


## Introduction to Flink in 30 minutes
https://www.youtube.com/watch?v=RCP9-HdId9w
- Interesting example where system needs to alert when customers consumes 90% of their quota
- Demonstration how to do batch on kafka
  - you can configure start and end kafka offset in flink as time, and this is how Flink knows that there won't be new data in batch processing
  - this may be nice for testing jobs on kafka
- Demonstration how to do streaming, Flink SQL, how to mix FQL with DataFrame, how to trigger alerts

## Mastering Flink on Kubernetes | Best Practices for Running Flink on Kubernetes
https://www.youtube.com/watch?v=JilAIFFzAxs

- Overview of challenges with running Flink on K8s
- One customer wanted to have flink with failower region, this ment they develop new operator, that was triggering snapshots on region-a,
  and jobs on region-b were configure to read from this snapshot. Snapshot was stored on s3.
- Similar trick was done for zero-downtime deployments of new version of job, first k8s orchiestrator triggers current job to do snapshot, and new job in new cluster would start from this snapshot.
  Trick was to also intrduce graceful period, between jobs shutdown, and new cluster start, which meant douplicates.
  Deduplication was done outside of Flink cluster, using fast KV store.

- Challenges with cross region faillover was that kafka in region-b may not been replicated, so when failore job resume from snapshot, kafka offset may not yet exists.
  They had do solve it by reading checkpoints, and find one that had offset that they can use.


## Getting started with Flink on k8s
https://nightlies.apache.org/flink/flink-kubernetes-operator-docs-main/docs/try-flink-kubernetes-operator/quick-start/

brew install kubectl
brew install minikube
brew install k9s
brew install helm

### deploy operator
minikube start


kubectl create -f https://github.com/jetstack/cert-manager/releases/download/v1.8.2/cert-manager.yaml


export STABLE_FLINK_OPERATOR_VERSION=1.9.0

- to find version go to https://flink.apache.org/downloads/#apache-flink-kubernetes-operator

helm repo add flink-operator-repo "https://downloads.apache.org/flink/flink-kubernetes-operator-${STABLE_FLINK_OPERATOR_VERSION}/"
helm install flink-kubernetes-operator flink-operator-repo/flink-kubernetes-operator


### submit job
https GET https://raw.githubusercontent.com/apache/flink-kubernetes-operator/release-1.8/examples/basic.yaml

kubectl create -f https://raw.githubusercontent.com/apache/flink-kubernetes-operator/release-1.8/examples/basic.yaml
kubectl logs -f deploy/basic-example


kubectl port-forward svc/basic-example-rest 8081

kubectl delete flinkdeployment/basic-example

### delete job
kubectl delete flinkdeployment/basic-example