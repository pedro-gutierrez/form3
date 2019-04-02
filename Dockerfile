FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git gcc libc-dev
ENV GOBIN /go/bin
WORKDIR $GOPATH/src/github.com/pedro-gutierrez/form3
ADD . . 
WORKDIR $GOPATH/src/github.com/pedro-gutierrez/form3/cmd
RUN go get
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/form3

FROM golang:alpine 
COPY --from=builder /go/bin/form3 /usr/local/bin/form3
RUN mkdir -p /etc/form3/schema
COPY --from=builder $GOPATH/src/github.com/pedro-gutierrez/form3/schema/* /etc/form3/schema/
CMD ["/usr/local/bin/form3", "--metrics=true", "--repo-migrations=/etc/form3/schema"]
