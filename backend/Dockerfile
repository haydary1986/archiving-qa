FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod ./
RUN go mod download || true

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Runtime
FROM alpine:3.19

RUN apk add --no-cache \
    ca-certificates \
    tesseract-ocr \
    tesseract-ocr-data-ara \
    tesseract-ocr-data-eng \
    ghostscript \
    imagemagick \
    tzdata

WORKDIR /app
COPY --from=builder /server .

EXPOSE 8080

CMD ["./server"]
