FROM golang:1.15.4

RUN go get golang.org/x/lint/golint

RUN go get github.com/julienschmidt/httprouter
RUN go get github.com/go-sql-driver/mysql

EXPOSE 8100 8200
