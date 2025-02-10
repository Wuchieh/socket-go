package socket

import (
	"encoding/json"
)

type _req struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func createContext(s *Socket, m *Member) *Context {
	c := new(Context)
	c.reset()
	c.m = m
	c.s = s
	return c
}

func handlerError(s *Socket, m *Member, err error) {
	if s == nil || m == nil {
		return
	}
	c := createContext(s, m)
	c.handlers = s.onError
	c.Data = err
	c.Next()
}

func handlerMessage(s *Socket, m *Member, b []byte) {
	if s == nil || m == nil {
		return
	}

	var req _req
	err := json.Unmarshal(b, &req)
	if err != nil {
		logf(err.Error())
		return
	}

	handlers, ok := s.handlers[req.Event]
	if !ok {
		handlerOtherEvent(s, m, req.Event, req.Data)
		return
	}

	c := createContext(s, m)
	c.Event = req.Event
	c.handlers = handlers
	c.Data = req.Data
	c.Next()
}

func handlerOtherEvent(s *Socket, m *Member, event string, data any) {
	c := createContext(s, m)
	c.Event = event
	c.Data = data
	c.Next()
}

func handlerOnConnect(s *Socket, m *Member) {
	if len(s.onConnect) == 0 {
		return
	}

	c := createContext(s, m)
	c.handlers = s.onConnect
	c.Next()
}

func handlerOnDisconnect(s *Socket, m *Member) {
	if len(s.onDisconnect) == 0 {
		return
	}

	c := createContext(s, m)
	c.handlers = s.onDisconnect
	c.Next()
}
