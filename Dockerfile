FROM golang:latest as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main/*.go ./main/
COPY aws/*.go ./aws/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./main/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/main /app/
CMD ["/app/main"]