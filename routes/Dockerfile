FROM golang:1.16.0
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o main .
EXPOSE 5432
CMD ["app/main"]