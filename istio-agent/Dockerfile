# Use a minimal base image with Go installed
FROM golang:1.22.4-alpine

# Set the working directory inside the container
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN go mod download

COPY . .


RUN go build -o main .

RUN chmod +x config-transformer.sh

CMD ["./main"]