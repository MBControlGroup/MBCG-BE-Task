FROM golang:1.8
RUN mkdir -p /go/src/task
ADD . /go/src/task
WORKDIR /go/src/task
CMD ["go", "run", "main.go"]