#build stage

FROM golang:1.22.0
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 9621
CMD ["app"]
