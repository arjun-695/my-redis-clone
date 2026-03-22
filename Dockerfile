# GO Server
# using golang alpine
FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o redis-clone .

#Run the server
FROM alpine:latest

WORKDIR /app

#copying final binary 
COPY --from=builder /app/redis-clone .

EXPOSE 5001

CMD ["./redis-clone"]