1. 怎麼啟動
開啟 terminal 後，cd 到此資料夾目錄後，執行 `go run .`
如果遇到 port 8080 被佔用，可以使用 `PORT=8081 go run .`。8081 為範例，可以改成其他 port 號碼。

Docker 啟動
1. 建置 Docker 映像檔
`docker build -t todo .`
2. 啟動容器
`docker run -p 8080:8080 todo`
3. 停止容器
`docker stop todo`
4. 刪除容器
`docker rm todo`
5. 刪除映像檔
`docker rmi todo`

## 1. 怎麼啟動

需要的環境（二選一）：

- 用 GO 執行：Go 1.22 以上
- 用 Docker

### 用 GO 執行
開啟 terminal 後，cd 到此資料夾目錄後，執行：
```bash
go run .
```

如果遇到 port 8080 被佔用，可以使用 
```bash
PORT=[自定義的 port 號碼] go run .
```
來改成其他 port 號。

### 用 Docker 執行（不需要安裝 Go）

```bash
docker build -t todo .
docker run --rm -p 8080:8080 todo
```

按 `Ctrl+C` 結束，container 會自動刪除，結束時會先處理完進行中的請求。

如果 port 8080 被佔用，改 `-p` 左邊的數字，例如 `-p 8081:8080`，之後連線到 `localhost:8081`。

不再需要映像檔時：

```bash
docker rmi todo
```

## 2. curl 範例
