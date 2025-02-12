package main

import (
	"github.com/wuchieh/socket-go"
	"net/http"
)

var (
	s *socket.Socket
)

func init() {
	s = socket.Default()

	s.On("echo", func(c *socket.Context) {
		c.Emit(c.Event, c.Data)
	})
}

func main() {
	http.HandleFunc("/ws", s.Handler)

	http.ListenAndServe(":8080", nil)
}
