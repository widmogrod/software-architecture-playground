## How to run tests?
Those tests are specific ones, and they require external dependencies to run. 
The Simplest way to do it is to use Docker-compose.

```
docker-compose up
go test ./... -i-exec-docker-compose-up
```
