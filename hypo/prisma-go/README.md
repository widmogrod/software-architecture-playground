# Introduction

Having fun with prisma-go and local development in kubernetes

```
alias k=kubectl

minikube start
minikube addons enable ingress
eval $(minikube docker-env)

minikube dashbord
minikube tunnel
```

## Local development
## Manual workflow
```
docker build . -t localhost/ghapp:0.0.3 -t localhost/ghapp:lates
# update container image version in deployment/k8s-native/app2.yaml 
k apply -f deployment/k8s-native/app2.yaml 

k port-forward svc/gh-app-svc 9999:9999
k delete -f deployment/k8s-native/app2.yaml
```

## skaffold
```
brew install skaffold
skaffold dev --port-forward
```

- Skaffold is much faster to start than devspace.sh
- Updates changes to any configuration automatically, so development in container is possible with only changed files sync. 
  In this project I use buildpacks and docker, and buildpack expirience provide reloading golang code withtou rebuilding after sync,
  and this feedback loop is quite acceptable
- Skaffold don't provide option to ssh to pod, so developer has to get pod `k get pod` and creat interactive session i.e. `k exec gh-app-7bbdd5cfc5-42txc -it -- sh` 
  which is a little troublesome, much better experience in this regard offers devspaces dev, where after everythig is setup, you're logged in to pod


## devspace.sh
```
devspaces dev
```

Devspace has quite interesting capabilities that make local workflow fantastic:
- ssh to pod just right after `devspace dev`
- sync files with the pod
- port-forward services
- auto-restart pod process

Such workflow is fast when you develop application, instead of k8s configuration

## Skaffold vs devspace.sh
- Syncing and reloading application seams faster for devspace, but sill in case of golang it takes time :/
- Devspace can open terminal and start forwarding after runing `devspace dev` command, Skaffold requires to provide --port-forward flag, and ssh to pod are next two steps
- Skaffold use buildpacks.io (Devspace don't have it, bit can be addedd as custom build option) and ability of just 
  building project without writing dockerfile manually push more things to runtime

# Using prisma
- Using prisma is quite strength forward, generation of SDK with types gives autocomplete that makes typing queries fast
- Generation of migrations and tooling to apply them is also well integrated
- In contrasts with ORMs of the past, Prisma don't have lazy resolving relations, which IMO is fantastic,
  no need to worry that to many data will be selected, without user knowing. Thanks to that it can be more manageable. 
- Prisma has also driver that in serverless setup can be deployed as proxy that manage connection pool

Now what I'm missing to complete development workflow is generation of APIs, gRPC is natural fit, 
but it will result in two technologies not leveraging information about types.

There is also other contender, GraphQL. GQL defines data model like `prisma`, but it also defines queries, mutations, subscriptions, that `gRPC gives`, 
but thanks to everything being same `intermediate mode` give possibility to have end-to-end types.

https://fauna.com/ is closes that I found to making