FROM golang:latest AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /go-recipe-srv

FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=build /go-recipe-srv /go-recipe-srv
EXPOSE 5500
USER nonroot:nonroot
ENTRYPOINT ["/go-recipe-srv"]