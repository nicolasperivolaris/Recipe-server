FROM golang:last
WORKDIR /go
COPY . .
RUN go mod download
RUN go build main.go
EXPOSE 5500
CMD [main]