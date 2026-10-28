## 1. 怎麼啟動

需要的環境（二選一）：

- 用 GO 執行：Go 1.24 以上
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

### 用 MySQL 存資料（選用，需要 Docker）

沒有設定 `MYSQL_DSN` 時，資料存在記憶體裡，server 一重開資料就不見了。設定 `MYSQL_DSN` 之後，資料會存進 MySQL，重開也還在。

1. 用 Docker 開一個 MySQL：

```bash
docker run -d --name todo-mysql -e MYSQL_ROOT_PASSWORD=secret -e MYSQL_DATABASE=todo -p 3306:3306 mysql:8.4
```

第一次啟動要等大約 20 秒。執行 `docker logs todo-mysql`，看到 `ready for connections` 而且同一行有 `port: 3306` 才算好。

2. 建立 todos 表格（`<` 是把檔案內容交給 container 裡的 mysql 執行）：

```bash
docker exec -i todo-mysql mysql -uroot -psecret todo < sql/todos_create_table.sql
```

3. 啟動 server：

```bash
MYSQL_DSN='root:secret@tcp(127.0.0.1:3306)/todo?parseTime=true' go run .
```

`parseTime=true` 不能省略，少了它讀取 `created_at` 時會出錯。

### 用 Docker 執行（不需要安裝 Go）

```bash
docker build -t todo .
docker run --rm -p 8080:8080 todo
```

按 `Ctrl+C` 結束，container 會自動刪除，結束時會先處理完進行中的請求。

如果 port 8080 被佔用，改 `-p` 左邊的數字，例如 `-p 8081:8080`，之後連線到 `localhost:8081`。

這種方式資料存在記憶體裡。要用 MySQL 的話，改用下面的 docker compose。

不再需要映像檔時：

```bash
docker rmi todo
```

### 用 docker compose 執行（Go + MySQL，不需要安裝 Go）

一個指令同時開兩個 container：一個跑 go 程式，一個跑 MySQL。資料存在 MySQL 裡，重開也還在。

```bash
docker compose up -d
```

MySQL 第一次啟動時，會自動執行 `sql/` 裡的檔案建好表格。go 程式會等 MySQL 準備好才啟動，所以第一次要等幾秒，之後就能連到 `localhost:8080`。

改了程式之後，要加 `--build` 重新建立映像檔，不然跑的還是舊的程式：

```bash
docker compose up -d --build
```

看 go 程式的 log：

```bash
docker compose logs app
```

關掉（container 會刪掉，但資料會留著，下次 `up` 資料還在）：

```bash
docker compose down
```

連資料一起清掉，下次 `up` 會從空的表格開始：

```bash
docker compose down -v
```

如果 port 8080 被佔用，把 `compose.yaml` 裡 `"8080:8080"` 左邊的數字改掉，例如 `"8081:8080"`。

### 跑測試

```bash
go test ./...
```

MySQL 的測試要用另一個資料庫 `todo_test`，因為每個測試開始前都會清空表格。沒有設定 `MYSQL_TEST_DSN` 時，MySQL 的測試會自動略過（SKIP）。

建立測試用的資料庫和表格（MySQL container 要先開好）：

```bash
docker exec todo-mysql mysql -uroot -psecret -e "CREATE DATABASE todo_test"
```

```bash
docker exec -i todo-mysql mysql -uroot -psecret todo_test < sql/todos_create_table.sql
```

連同 MySQL 的測試一起跑：

```bash
MYSQL_TEST_DSN='root:secret@tcp(127.0.0.1:3306)/todo_test?parseTime=true' go test ./...
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

以上是存在記憶體時的回應。用 MySQL 時，`created_at` 只到秒，而且是 UTC 時間，例如 `"2026-10-14T01:39:45Z"`。



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


| 情況                           | 狀態碼 | 回應                                |
| ---------------------------- | --- | --------------------------------- |
| id 不是正整數，例如 `/todos/abc`     | 400 | `{"error":"invalid id"}`          |
| `done` 參數不是 `true` / `false` | 400 | `{"error":"invalid done"}`        |
| JSON 格式錯誤，或帶了不認識的欄位          | 400 | `{"error":"invalid JSON"}`        |
| title 是空白，或超過 100 個字         | 400 | `{"error":"invalid title"}`       |
| PUT 時 `title` 和 `done` 都沒傳   | 400 | `{"error":"no fields to update"}` |
| 找不到該 id 的 todo               | 404 | `{"error":"todo not found"}`      |


例如：

```bash
curl -i localhost:8080/todos/999
```

```
HTTP/1.1 404 Not Found
Content-Type: application/json

{"error":"todo not found"}
```



## 3. 做了哪些加分題

- Graceful shutdown
在 server.go 中，按 Ctrl+C 後不再接新請求，等處理中的請求完成才結束，最多等 5 秒。
- Logging middleware
在 middleware.go 中，每個請求印出 method、path、狀態碼和處理時間。
- 背景工作
在 reporter.go 中，使用 time.Ticker 每 10 秒印出一次未完成的 todo 數量。
- Dockerfile
使用 multi-stage build，最後的映像檔只放編譯好的執行檔。



## 4. 遇到的困難，或你覺得還可以改進的地方



### 遇到的困難

- C# 的習慣：我常常不自覺把 struct 當成 C# 的 class 來想。像是用一個 nil 的指標去呼叫方法，我以為拿不到正常的結果，結果只要方法裡沒有用到欄位，就能正常執行。後來才知道 go 傳參數都是複製一份，要改到原本的值就要傳指標，所以 MemoryStore 的方法都是用 *MemoryStore。
- interface：go 不用像 C# 一樣宣告實作哪個介面，方法都有就算實作，也沒有繼承。一開始對 ResponseWriter 不熟，不知道裡面有哪些方法。看 middleware 的時候，我把 statusRecorder 的 WriteHeader 刪掉，程式也沒有報錯，也找不到哪裡有呼叫它，那時候完全看不懂。後來才理解 statusRecorder 用的是嵌入：把 ResponseWriter 直接放進 struct，它的方法就會變成 statusRecorder 的方法，我自己寫的 WriteHeader 會蓋掉原本的。所以刪掉之後程式還是能跑，只是 log 印出來的狀態碼會一直是 200。這個模式習慣 C# 的我有點不直觀。
- slice：一開始覺得很像 C# 的 List，但又不太像。我以為把 slice 傳進函式 append 之後，外面的長度也會跟著變，結果沒有。後來才懂 slice 裡面是 ptr、len、cap，傳進去時 ptr 指向同一個 array，所以改元素外面看得到，但 len 和 cap 是各自的。
- 併發：寫 Get 的時候我沒有上鎖，因為覺得只是讀資料，不會改到東西。後來跑 go test -race 才發現，Get 在讀 map 的時候，Create 可能同時在寫，server 真的遇到會直接當掉。這讓我知道只要有人可能在寫，讀也要上鎖。後來 MemoryStore 改成 RWMutex，讀和讀可以同時進行，但有人在寫的時候就要等。
- 方法要寫在哪裡：原本 handler 的路由都寫在 New 裡面，我想把每個路由拆成獨立的函式，就寫了 func (s *store.Store) list(...)，想讓它直接用 store，結果編譯失敗。後來才知道不能幫別的 package 的型別加方法，interface 也不能有實作，所以改成在 handler 自己定義一個 api struct，把 store 放進去，方法掛在 api 上。
- channel：這是我最不熟的部分。我以為 close 就是把 channel 釋放掉，後來才知道它只是告訴接收的一方不會再送了，裡面還沒讀的值還是讀得到。容量也搞混過，以為 make(chan int, 3) 只放了兩個值的話，第三格會讀到 0。
- 測試：之前做 Unity 沒有寫過單元測試，一開始不知道 _test.go 要怎麼寫。這次寫下來的感覺是先設定好情境，和該情境應該產生的結果。但自己寫程式又自己寫測試，有點球員兼裁判：第一次用批改程式檢查時，有 4 個規格要求的情況（像是 id 是 0 要回 400），我的測試都沒有測到。後來把這 4 個情況也補進測試裡。
- defer 什麼時候執行：加 MySQL 時，同一個錯誤我犯了兩次。我在 NewMySQLStore 裡寫了 defer db.Close()，以為程式結束才會關，結果 defer 是在 NewMySQLStore 自己結束時就執行，拿到的連線已經被關掉，每次查詢都回 database is closed。後來寫測試用的輔助函式又犯了一次。另外在 Update 裡，我在 if 裡面用 := 想設定外面的變數，結果是建立了一個新的變數，外面的還是空的。
- MySQL 的併發：一次只送一個請求時都正常，同時很多請求就出問題。第一個是兩個 Update 同時改同一筆時，兩邊都先讀出舊資料再寫回去，後寫的會把先寫的蓋掉，AI 幫我用測試同時跑了 200 次，每次都發生。解法是交易（transaction）加上 SELECT ... FOR UPDATE，讓同一筆資料在改完之前，其他請求要先等。這段我一開始看不懂，請 AI 先示範一次，再一步一步看懂流程。第二個是把 MemoryStore 的併發測試搬到 MySQL 時出現 Too many connections：sql.DB 預設不限制連線數，300 個 goroutine 同時開連線，超過了 MySQL 的上限 151 條，最後用 SetMaxOpenConns 限制最多 10 條，多的請求排隊等。
- docker compose：compose.yaml 的參數我還不熟，是請 AI 寫好，再一行一行對照我之前用過的 docker run 指令看懂。裡面有兩個重點。第一個是在 container 裡，127.0.0.1 指的是 container 自己，所以 go 程式要用服務名稱 db 去連 MySQL。第二個是啟動順序：先只寫 depends_on，從空的狀態啟動時，go 程式比 MySQL 早了 5 秒去連線，連不上就結束了。depends_on 只保證 MySQL 先啟動，不保證它已經準備好。後來在 MySQL 加上 healthcheck（每 2 秒試著連一次），go 程式等檢查通過才啟動。

### 還可以改進的地方

- 熟悉 go 的 package：我有程式的基本概念，卡住的通常是不知道要用哪個 package、呼叫哪個函式。這次主要是先問 AI 該用什麼，再自己寫、跑測試確認，也請 AI 出題考我，用測驗找出自己不懂的地方。之後想繼續用這種方式，把常用的 struct 和 api 熟悉起來。
- 500 沒有測到：handler 的測試覆蓋率是 94%，沒測到的是 store 出錯時回 500 的那段。MemoryStore 不會回傳這種錯誤，所以之後想寫一個故意出錯的假 store 來測。
- MySQL 的重複程式碼：mysql.go 裡把一筆資料讀進 Todo 的 Scan 寫了三次，之後想抽成一個共用的函式。mysql_test.go 是把 memory_test.go 的測試逐個複製過來，同樣的測試案例有兩份，之後想改成兩個 store 共用同一組測試。
- 密碼寫在檔案裡：MySQL 的密碼直接寫在 compose.yaml 裡，用指令連 MySQL 時也會看到 `Using a password on the command line interface can be insecure` 的警告。之後想把密碼移到另一個不放進 git 的設定檔（.env）。
