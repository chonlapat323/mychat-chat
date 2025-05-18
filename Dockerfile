FROM golang:1.24

WORKDIR /app
COPY . .

RUN go build -o mychat-room .

EXPOSE 5001

CMD ["./mychat-room"]
