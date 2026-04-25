# ============================================================
# Go API 多阶段构建 - 最小化镜像
# ============================================================
FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOMAXPROCS=1 CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/api-server .

# ==================== 运行阶段 ====================
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /build/api-server /usr/local/bin/api-server

EXPOSE 8967

CMD ["api-server"]
