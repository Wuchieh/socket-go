package main

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/wuchieh/socket-go"
	"log"
	"net/http"
	"slices"
	"time"
)

type H map[string]any

var (
	s *socket.Socket
)

func getName(c *socket.Context) string {
	m := c.GetMember()
	val, ok := m.Values.Load("name")
	if !ok {
		return ""
	}
	name, ok := val.(string)
	if !ok {
		return ""
	}
	return name
}

func setName(c *socket.Context, name string) {
	c.GetMember().Values.Store("name", name)
}

func getRoom(c *socket.Context) string {
	rooms := c.GetMember().GetRooms()
	if len(rooms) > 0 {
		return rooms[0]
	}
	return ""
}

func init() {
	s = socket.Default()

	s.On("echo", func(c *socket.Context) {
		c.Emit(c.Event, c.Data)
	})

	s.On("set_name", func(c *socket.Context) {
		name, ok := c.Data.(string)
		if ok {
			setName(c, name)
		}
	})

	s.On("join_room", func(c *socket.Context) {
		room, ok := c.Data.(string)
		if ok {
			rooms := c.GetMember().GetRooms()
			switch room {
			case "room_1":
				if slices.Contains(rooms, "room_2") {
					c.Leave("room_2")
				}
				c.Join(room)
			case "room_2":
				if slices.Contains(rooms, "room_1") {
					c.Leave("room_1")
				}
			default:
				return
			}
			c.Join(room)
			c.Emit("join_room", room)
		}
	})

	s.OnError(func(c *socket.Context) {
		err := c.Data.(error)
		cErr := new(websocket.CloseError)
		if errors.As(err, &cErr) {
			if cErr.Code == 1001 {
				return
			}
		}

		log.Println(err)
	})

	s.On("send_massage", func(c *socket.Context) {
		message, ok := c.Data.(string)
		if !ok {
			log.Println("not found message")
			return
		}

		name := getName(c)
		if name == "" {
			log.Println("not found name")
			return
		}

		room := getRoom(c)
		if room == "" {
			log.Println("not found room")
			return
		}

		err := c.To(room).Emit("send_massage", H{"name": name, "message": message, "time": time.Now()})
		if err != nil {
			log.Println("emit error:", err)
			return
		}
		log.Println("send_massage:", message)

		//err := s.Emit("send_massage", H{"name": name, "message": message, "time": time.Now()})
		//if err != nil {
		//	log.Println("emit error:", err)
		//	return
		//}

		//err := c.Emit("send_massage", H{"name": name, "message": message, "time": time.Now()})
		//if err != nil {
		//	log.Println("emit error:", err)
		//	return
		//}
	})
}

func main() {
	http.HandleFunc("/ws", s.Handler)

	http.ListenAndServe(":8080", nil)
}
