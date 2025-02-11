package main

import (
	"fmt"
	"github.com/wuchieh/socket-go"
	"log"
	"net/http"
)

var (
	s *socket.Socket
)

func init() {
	s = socket.Default()

	s.On("echo", func(c *socket.Context) {
		fmt.Println(c.Data)
		c.Emit("message", c.Data)
	})

	s.On("add", func(c *socket.Context) {
		m := c.GetMember()
		actual, loaded := m.Values.LoadOrStore("count", 0)
		if !loaded {
			c.Emit("message", actual)
			return
		}

		ai := actual.(int)
		ai++
		m.Values.Store("count", ai)
		c.Emit("message", ai)
	})

	s.On("join", func(c *socket.Context) {
		room, ok := c.Data.(string)
		if !ok {
			c.Emit("error", "room name need is string")
			return
		}
		c.Join(room)
		c.BroadcastTo("member_join", nil)
	})

	s.On("leave", func(c *socket.Context) {
		room, ok := c.Data.(string)
		if !ok {
			c.Emit("error", "room name need is string")
			return
		}
		c.Leave(room)
		c.BroadcastTo("member_leave", nil)
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
