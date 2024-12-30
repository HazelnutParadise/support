FROM golang:1.23-alpine AS builder

# 設定工作目錄
WORKDIR /app

# 複製當前目錄下的所有檔案到容器的工作目錄
COPY . .

# 執行編譯命令，生成二進位檔案 myapp
RUN go build -o app

# 第二階段：使用更小的基礎映像來建立最終的映像
FROM scratch

# 設定工作目錄
WORKDIR /app

# 從 builder 階段複製編譯後的二進位檔案到當前階段的工作目錄
COPY --from=builder /app/app .
COPY --from=builder /app/templates /app/templates

# 設定容器啟動時執行的命令
CMD ["/app/app"]