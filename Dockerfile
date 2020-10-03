FROM golang:1.14 AS builder
WORKDIR /go/src/app
COPY . .
RUN \
  GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64 \
  go build -a -installsuffix cgo -o /bin/reconfigure-storage-class-annotation .

FROM scratch
COPY --from=builder /bin/reconfigure-storage-class-annotation /bin/reconfigure-storage-class-annotation
CMD ["/bin/reconfigure-storage-class-annotation"]