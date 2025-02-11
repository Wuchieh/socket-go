# Socket-go

簡易socket 框架

輸入以及輸出的數據皆為JSON結構化數據

## 安裝

```shell
  go get github.com/wuchieh/socket-go@latest
```

## 使用

[簡易的使用範例](example/sample/main.go)

```go
package main

import (
	"fmt"
	socket "github.com/wuchieh/socket-go"
	"net/http"
)

var s *socket.Socket

func init() {
	// 初始化 socket
	s = socket.Default()

	// 加入一個事件
	s.On("echo", func(c *socket.Context) {
		fmt.Println(c.Data)
		c.Emit("message", c.Data)
	})
}

func main() {
	// 綁定 /ws 路由
	http.HandleFunc("/ws", s.Handler)

	// 啟動伺服器
	http.ListenAndServe(":8080", nil)
}
```

## Emit 發送事件

- Socket 使用時為群體廣播
- Context 使用時只會回傳給觸發此事件的 Member
- Member 使用時只會回傳給此 Member
- ContextTo 使用時 會先取得 To(room) 的 Member，再將 Except(room) 的 Member 移除，最後再發送訊息
- BroadcastTo 除了 Except(Member) 所有人都會收到消息

## TODO
- [x] 新增 Broadcast 方法
- [ ] 新增 socket-go 的 npm 套件
- [ ] 加入 socket-go 的 go client 端
- [ ] 優化 example