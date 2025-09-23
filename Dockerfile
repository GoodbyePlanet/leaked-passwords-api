FROM golang:1.25.1-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o leaked-passwords-api ./src
EXPOSE 8080
CMD ["./leaked-passwords-api"]
