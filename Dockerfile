# Build image builder
FROM golang:alpine as builder
WORKDIR /home/api
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o app

# Build app image
FROM alpine:latest as image
WORKDIR /home/api
COPY --from=builder /home/api/app .
CMD [ "./app" ]