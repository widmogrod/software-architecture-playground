FROM golang:1.17-buster AS builder

WORKDIR /app
COPY migrations .
COPY main.go ./
COPY dev.db ./
COPY schema.prisma ./
COPY go.mod go.sum ./

RUN go mod download
RUN go run github.com/prisma/prisma-client-go prefetch

ENV APP_PORT=8080
#ENV CGO_ENABLED=0
ENV GOARCH=arm64
ENV GOOS=linux

RUN go run github.com/prisma/prisma-client-go generate
RUN go run github.com/prisma/prisma-client-go migrate deploy
#RUN go build -a -installsuffix cgo -o main .
RUN go build -o main .


FROM golang:1.17-buster
WORKDIR /opt/app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/main .
COPY --from=builder /app/dev.db .

CMD ["./main"]