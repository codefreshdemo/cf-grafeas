FROM golang:1.8.3 AS builder
WORKDIR /go/src/github.com/codefresh/cf-grafeas
ADD . .
RUN go get -u github.com/golang/dep/cmd/dep && dep ensure -v
RUN CGO_ENABLED=0 GOOS=linux go build -o cf-grafeas main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
EXPOSE 8091
COPY --from=builder /go/src/github.com/codefresh/cf-grafeas/cf-grafeas .
CMD ["./cf-grafeas"]
