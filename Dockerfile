FROM golang:1.23-alpine as builder

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN go build



FROM scratch
WORKDIR /app

COPY --from=builder /app/vfio-device-plugin .

CMD ["/app/vfio-device-plugin"]
