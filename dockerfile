FROM golang:1.9.2
COPY . /go/src/github.com/AntsEclipse/gochat
WORKDIR /go/src/github.com/AntsEclipse/gochat

RUN cd auth && go build .
RUN cd register && go build .
RUN cd user && go build .
