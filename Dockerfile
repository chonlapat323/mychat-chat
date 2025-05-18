FROM golang:1.21-alpine AS builder

# ทำให้ build fail ไว ถ้ามีปัญหา
RUN apk add --no-cache git

WORKDIR /app

# ดึง dependency ก่อน
COPY go.mod go.sum ./
RUN go mod download

# คัดลอก source code
COPY . .

# Compile binary
RUN go build -o main .

# Final stage: minimal runtime image
FROM alpine:latest

WORKDIR /app

# เพิ่ม timezone (optional)
RUN apk add --no-cache tzdata

# คัดลอก binary จาก build stage
COPY --from=builder /app/main .


EXPOSE 5001

# CMD
CMD ["./main"]
