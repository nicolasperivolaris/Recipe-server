FROM golang:last
COPY . /go
WORKDIR /go
RUN go build main.go
EXPOSE 5500