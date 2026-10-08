# ---------- 第 1 階段：編譯 ----------
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /todo .

# ---------- 第 2 階段：只放執行檔 ----------
FROM scratch
COPY --from=build /todo /todo
EXPOSE 8080
ENTRYPOINT ["/todo"]