package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/wuchieh/socket-go"
	"log"
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

	s.On("bind", func(c *socket.Context) {
		type req struct {
			Username string `json:"username"`
		}

		var r req
		if err := c.Bind(&r); err != nil {
			log.Println(err)
		}

		fmt.Printf("%#v\n", r)
	})

	s.On("chat", func(c *socket.Context) {
		var t *socket.ContextTo
		for i, room := range c.GetMember().GetRooms() {
			if i == 0 {
				t = c.To(room)
			} else {
				t = t.To(room)
			}
		}

		if t != nil {
			t.Emit("message", c.Data)
		}
	})
}

func main() {
	http.HandleFunc("/ws", s.Handler)

	http.ListenAndServe(":8080", nil)
}
