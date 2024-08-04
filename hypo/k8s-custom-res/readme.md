

```
https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/

brew install kubectl
brew install minikube
brew install k9s
brew install helm

colima start --edit
kubernetes:
  enable: true
```


```
minikube start
kubectl get po -A

alias k=kubectl
```

```
https://debezium.io/documentation/reference/stable/operations/kubernetes.html

 minikube addons enable registry
 kubectl create ns debezium-example
 brew install operator-sdk
 operator-sdk olm install
 
 kubectl create -f https://operatorhub.io/install/strimzi-kafka-operator.yaml

# username=debezium
# password=dbz
 cat << EOF | kubectl create -n debezium-example -f -
apiVersion: v1
kind: Secret
metadata:
  name: debezium-secret
  namespace: debezium-example
type: Opaque
data:
  username: ZGViZXppdW0=
  password: ZGJ6
EOF

cat << EOF | kubectl create -n debezium-example -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: connector-configuration-role
  namespace: debezium-example
rules:
- apiGroups: [""]
  resources: ["secrets"]
  resourceNames: ["debezium-secret"]
  verbs: ["get"]
EOF

cat << EOF | kubectl create -n debezium-example -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: connector-configuration-role-binding
  namespace: debezium-example
subjects:
- kind: ServiceAccount
  name: debezium-connect-cluster-connect
  namespace: debezium-example
roleRef:
  kind: Role
  name: connector-configuration-role
  apiGroup: rbac.authorization.k8s.io
EOF

cat << EOF | kubectl create -n debezium-example -f -
apiVersion: kafka.strimzi.io/v1beta2
kind: Kafka
metadata:
  name: debezium-cluster
spec:
  kafka:
    version: 3.7.1
    replicas: 1
    listeners:
      - name: plain
        port: 9092
        type: internal
        tls: false
      - name: tls
        port: 9093
        type: internal
        tls: true
        authentication:
          type: tls
      - name: external
        port: 9094
        type: nodeport
        tls: false
    storage:
      type: jbod
      volumes:
      - id: 0
        type: persistent-claim
        size: 10Gi
        deleteClaim: false
    config:
      offsets.topic.replication.factor: 1
      transaction.state.log.replication.factor: 1
      transaction.state.log.min.isr: 1
      default.replication.factor: 1
      min.insync.replicas: 1
  zookeeper:
    replicas: 1
    storage:
      type: persistent-claim
      size: 10Gi
      deleteClaim: false
  entityOperator:
    topicOperator: {}
    userOperator: {}
EOF

kubectl wait kafka/debezium-cluster --for=condition=Ready --timeout=300s -n debezium-example

cat << EOF | kubectl create -n debezium-example -f -
apiVersion: v1
kind: Service
metadata:
  name: mysql
spec:
  ports:
  - port: 3306
  selector:
    app: mysql
  clusterIP: None
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
spec:
  selector:
    matchLabels:
      app: mysql
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - image: quay.io/debezium/example-mysql:2.7
        name: mysql
        env:
        - name: MYSQL_ROOT_PASSWORD
          value: debezium
        - name: MYSQL_USER
          value: mysqluser
        - name: MYSQL_PASSWORD
          value: mysqlpw
        ports:
        - containerPort: 3306
          name: mysql
EOF

kubectl -n kube-system get svc registry -o jsonpath='{.spec.clusterIP}'
# 10.99.42.250

cat << EOF | kubectl create -n debezium-example -f -
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnect
metadata:
  name: debezium-connect-cluster
  annotations:
    strimzi.io/use-connector-resources: "true"
spec:
  version: 3.7.1
  replicas: 1
  bootstrapServers: debezium-cluster-kafka-bootstrap:9092
  config:
    config.providers: secrets
    config.providers.secrets.class: io.strimzi.kafka.KubernetesSecretConfigProvider
    group.id: connect-cluster
    offset.storage.topic: connect-cluster-offsets
    config.storage.topic: connect-cluster-configs
    status.storage.topic: connect-cluster-status
    # -1 means it will use the default replication factor configured in the broker
    config.storage.replication.factor: -1
    offset.storage.replication.factor: -1
    status.storage.replication.factor: -1
  build:
    output:
      type: docker
      image: 10.99.42.250/debezium-connect-custom:latest
    plugins:
      - name: debezium-mysql-connector
        artifacts:
          - type: tgz
            url: https://repo1.maven.org/maven2/io/debezium/debezium-connector-mysql/2.7.0.Final/debezium-connector-mysql-2.7.0.Final-plugin.tar.gz
      - name: debezium-mongodb-connector
        artifacts:
          - type: tgz
            url: https://repo1.maven.org/maven2/io/debezium/debezium-connector-mongodb/2.7.0.Final/debezium-connector-mongodb-2.7.0.Final-plugin.tar.gz
EOF


cat << EOF | kubectl create -n debezium-example -f -
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnector
metadata:
  name: debezium-connector-mysql
  labels:
    strimzi.io/cluster: debezium-connect-cluster
spec:
  class: io.debezium.connector.mysql.MySqlConnector
  tasksMax: 1
  config:
    tasks.max: 1
    database.hostname: mysql
    database.port: 3306
    database.user: \${secrets:debezium-example/debezium-secret:username}
    database.password: \${secrets:debezium-example/debezium-secret:password}
    database.server.id: 184054
    topic.prefix: mysql
    database.include.list: inventory
    schema.history.internal.kafka.bootstrap.servers: debezium-cluster-kafka-bootstrap:9092
    schema.history.internal.kafka.topic: schema-changes.inventory
EOF
```

```
kubectl run -n debezium-example -it --rm --image=mysql:8.2 --restart=Never --env MYSQL_ROOT_PASSWORD=debezium mysqlterm -- mysql -hmysql -P3306 -uroot -pdebezium
```

```
cat << EOF | kubectl create -n debezium-example -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: kafka-ui-config
  namespace: debezium-example
data:
  application.yml: |
    kafka:
      clusters:
        - name: debezium-cluster
          bootstrapServers: "debezium-cluster-kafka-bootstrap:9092"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kafka-ui
  namespace: debezium-example
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kafka-ui
  template:
    metadata:
      labels:
        app: kafka-ui
    spec:
      containers:
        - name: kafka-ui
          image: provectuslabs/kafka-ui:latest
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: config-volume
              mountPath: /config
          env:
            - name: KAFKA_CLUSTERS_FILE
              value: /config/application.yml
      volumes:
        - name: config-volume
          configMap:
            name: kafka-ui-config
---
apiVersion: v1
kind: Service
metadata:
  name: kafka-ui
  namespace: debezium-example
spec:
  selector:
    app: kafka-ui
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
      nodePort: 30000
  type: LoadBalancer
EOF

# to expose load balancer
 minikube tunnel

```

```
kubectl create ns debezium-ui

cat << EOF | kubectl create -n debezium-ui -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: debezium-ui
  namespace: debezium-ui
spec:
  replicas: 1
  selector:
    matchLabels:
      app: debezium-ui
  template:
    metadata:
      labels:
        app: debezium-ui
    spec:
      containers:
        - name: debezium-ui
          image: debezium/debezium-ui:2.1.2.Final
          ports:
            - containerPort: 8080
          env:
            - name: KAFKA_CONNECT_URIS
              value: "http://debezium-connect-cluster-connect-api.debezium-example.svc.cluster.local:8083"
---
apiVersion: v1
kind: Service
metadata:
  name: debezium-ui
  namespace: debezium-ui
spec:
  selector:
    app: debezium-ui
  ports:
    - protocol: TCP
      port: 8082
      targetPort: 8080
  type: LoadBalancer
EOF
```


```
helm repo add mongodb https://mongodb.github.io/helm-charts
helm install community-operator mongodb/community-operator -n debezium-example

cat << EOF | kubectl create -n debezium-example -f -
apiVersion: mongodbcommunity.mongodb.com/v1
kind: MongoDBCommunity
metadata:
  name: mongodb
spec:
  members: 3
  type: ReplicaSet
  version: "6.0.5"
  security:
    authentication:
      modes: ["SCRAM"]
  users:
    - name: debezium
      db: admin
      passwordSecretRef: # a reference to the secret that will be used to generate the user's password
        name: my-user-password
      roles:
        - name: clusterAdmin
          db: admin
        - name: userAdminAnyDatabase
          db: admin
      scramCredentialsSecretName: my-scram
  additionalMongodConfig:
    storage.wiredTiger.engineConfig.journalCompressor: zlib

# the user credentials will be generated from this secret
# once the credentials are generated, this secret is no longer required
---
apiVersion: v1
kind: Secret
metadata:
  name: my-user-password
type: Opaque
stringData:
  password: ZGJ6
EOF
```

https://github.com/mongodb/mongodb-kubernetes-operator/blob/master/docs/deploy-configure.md#deploy-a-replica-set

<metadata.name>-<auth-db>-<username>
kubectl get secret mongodb-admin-debezium -n debezium-example -o json

mongosh "mongodb+srv://debezium:ZGJ6@mongodb-svc.debezium-example.svc.cluster.local/admin?replicaSet=mongodb&ssl=false"

# http POST http://localhost:8083/connectors Content-Type:application/json < mongo-con.json

```bash
cat << EOF | kubectl create -n debezium-example -f -
apiVersion: kafka.strimzi.io/v1beta2
kind: KafkaConnector
metadata:
  name: debezium-connector-mongodb-2
  namespace: debezium-example
  labels:
    strimzi.io/cluster: debezium-connect-cluster
spec:
  class: io.debezium.connector.mongodb.MongoDbConnector
  tasksMax: 1
  config:
    mongodb.connection.string: "mongodb+srv://debezium:ZGJ6@mongodb-svc.debezium-example.svc.cluster.local/admin?replicaSet=mongodb&ssl=false"
    topic.prefix: mongo
    database.history.kafka.topic: "schema-changes.myMongoDB"
    database.include.list:  "inventory"
    schema.history.internal.kafka.bootstrap.servers: debezium-cluster-kafka-bootstrap:9092
    schema.history.internal.kafka.topic: schema-changes.admin
EOF
```