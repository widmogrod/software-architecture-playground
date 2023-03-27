#FROM golang:1.20 as builder
#
#WORKDIR /app
#COPY go.mod go.sum ./
#RUN go mod download
#COPY . .
#
#RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
EXPOSE 8080
#COPY --from=builder /app/main .
COPY main .
RUN uname -a
RUN apk add --update binutils curl
RUN readelf -h ./main
CMD ["./main"]