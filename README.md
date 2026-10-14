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

以下假設 server 跑在 `localhost:8080`，改過 port 的話請自行替換。
想看狀態碼的話，在 `curl` 後面加 `-i`。

### 新增 todo

```bash
curl -X POST localhost:8080/todos -H 'Content-Type: application/json' -d '{"title":"買牛奶"}'
```

回應 `201 Created`：

```json
{"id":1,"title":"買牛奶","done":false,"created_at":"2026-10-14T09:39:45.474691+08:00"}
```

### 列出全部 todo

```bash
curl localhost:8080/todos
```

回應 `200 OK`，依 id 由小到大排序，沒有資料時回傳 `[]`：

```json
[{"id":1,"title":"買牛奶","done":false,"created_at":"2026-10-14T09:39:45.474691+08:00"},{"id":2,"title":"寫作業","done":false,"created_at":"2026-10-14T09:39:45.487402+08:00"}]
```

### 只列出已完成 / 未完成的 todo

```bash
curl 'localhost:8080/todos?done=true'
```

```bash
curl 'localhost:8080/todos?done=false'
```

`done` 只接受 `true` 或 `false`。

### 取得單筆 todo

```bash
curl localhost:8080/todos/1
```

回應 `200 OK`：

```json
{"id":1,"title":"買牛奶","done":false,"created_at":"2026-10-14T09:39:45.474691+08:00"}
```

### 更新 todo

`title` 和 `done` 可以只傳其中一個，沒傳的欄位不會變。

```bash
curl -X PUT localhost:8080/todos/1 -H 'Content-Type: application/json' -d '{"done":true}'
```

```bash
curl -X PUT localhost:8080/todos/1 -H 'Content-Type: application/json' -d '{"title":"買豆漿","done":false}'
```

回應 `200 OK`，內容是更新後的 todo：

```json
{"id":1,"title":"買豆漿","done":false,"created_at":"2026-10-14T09:39:45.474691+08:00"}
```

### 刪除 todo

```bash
curl -i -X DELETE localhost:8080/todos/1
```

回應 `204 No Content`，沒有 body。

### 錯誤回應

錯誤一律回傳 `{"error": "訊息"}`：

| 情況 | 狀態碼 | 回應 |
| --- | --- | --- |
| id 不是正整數，例如 `/todos/abc` | 400 | `{"error":"invalid id"}` |
| `done` 參數不是 `true` / `false` | 400 | `{"error":"invalid done"}` |
| JSON 格式錯誤，或帶了不認識的欄位 | 400 | `{"error":"invalid JSON"}` |
| title 是空白，或超過 100 個字 | 400 | `{"error":"invalid title"}` |
| PUT 時 `title` 和 `done` 都沒傳 | 400 | `{"error":"no fields to update"}` |
| 找不到該 id 的 todo | 404 | `{"error":"todo not found"}` |

例如：

```bash
curl -i localhost:8080/todos/999
```

```
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"todo not found"}
```

