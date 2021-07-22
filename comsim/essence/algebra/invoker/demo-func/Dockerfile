FROM golang:1.16-alpine

WORKDIR /go/src/app
EXPOSE 9666
COPY . .

#RUN go get -d -v ./...
#RUN go install -v ./...

CMD ["go","run","main.go"]