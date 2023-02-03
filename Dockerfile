FROM golang:alpine

WORKDIR /app

COPY . .

RUN go build -o ./imgcutter ./cmd/main.go

EXPOSE 8080

CMD [ "./imgcutter" ]