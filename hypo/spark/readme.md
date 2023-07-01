# Having fun with Apache Spark
## Setup
```
conda activate software-architecture-playground
python3 -m pip install pyspark
brew install --cask adoptopenjdk/openjdk/adoptopenjdk8
export JAVA_HOME='/Library/Java/JavaVirtualMachines/adoptopenjdk-8.jdk/Contents/Home/'
```

## Run
```
spark-submit example1.py
```


# Minikube
https://jaceklaskowski.github.io/spark-kubernetes-book/demo/spark-shell-on-minikube/
```
docker context ls
colima start -c 8 -m 10 -d 50

minikube start --cpus 4 --memory 8192

alias k=kubectl


#spark build
 ./build/mvn -DskipTests -Pkubernetes clean package

```