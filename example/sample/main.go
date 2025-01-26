package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/wuchieh/socket-go"
	"net/http"
)

var (
	s *socket.Socket
)

func init() {
	s = socket.NewSocket()

	s.SetGrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})

	s.On("echo", func(c *socket.Context) {
		fmt.Println(c.Data)
		c.Emit("message", c.Data)
	})

	s.On("join", func(c *socket.Context) {
		c.Join("chatroom")
	})

	s.On("leave", func(c *socket.Context) {
		c.Leave("chatroom")
	})

	s.On("chat", func(c *socket.Context) {
		c.To("chatroom").Emit("message", c.Data)
	})
}

func main() {
	http.HandleFunc("/ws", s.Handler)

	http.ListenAndServe(":8080", nil)
}
