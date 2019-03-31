FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
ENV GOBIN /go/bin
WORKDIR $GOPATH/src/github.com/pedro-gutierrez/form3
COPY . .
WORKDIR $GOPATH/src/github.com/pedro-gutierrez/form3/cmd
RUN go get
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/form3

FROM scratch
COPY --from=builder /go/bin/form3 /usr/local/bin/form3
CMD ["/usr/local/bin/form3"]
