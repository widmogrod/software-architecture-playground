# Airflow
## Introduction
Having fun with Apache Airflow.

Hub of operators and examples https://registry.astronomer.io/

## Setup

Python setup
```
conda activate --stack software-architecture-playground 

```

```

./script/bootstrap
./script/test
```

```
colima start -m 7 -c 5

docker-compose up airflow-init
docker-compose up

# login & password -> airflow airflow
open localhost:8080

```