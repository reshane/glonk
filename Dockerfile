# Build binary
FROM golang:latest AS builder
WORKDIR /glonk
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/glonk

# Copy to image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /main .
COPY --from=builder ./glonk/static ./static
EXPOSE 8080
CMD ["./main"]
