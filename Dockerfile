FROM golang:1.18

WORKDIR /go/src/kickback-app/common

COPY . .

RUN go mod tidy

ENTRYPOINT ["go", "test", "-v", "./..."]