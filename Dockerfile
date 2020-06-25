FROM golang:1.14.4
WORKDIR /go/src/github.com/paulczar/m13k
ENV GO111MODULE=on
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o m13k main.go

FROM ubuntu:latest
RUN apt-get -yq update \
    && apt-get -yq install wget curl ca-certificates
RUN wget -O /usr/bin/ytt https://github.com/k14s/ytt/releases/download/v0.28.0/ytt-linux-amd64 \
    && chmod +x /usr/bin/ytt
WORKDIR /app/
COPY --from=0 go/src/github.com/paulczar/m13k/m13k .
USER 1001
CMD ["/app/m13k"]