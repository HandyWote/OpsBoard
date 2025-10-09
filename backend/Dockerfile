FROM docker.1ms.run/golang:1.24 AS builder

ENV GOPROXY=https://goproxy.cn

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM docker.1ms.run/golang:1.24

WORKDIR /app

COPY --from=builder /app/server /app/server

ENV PORT=9012

EXPOSE 9012

ENTRYPOINT ["./server"]
