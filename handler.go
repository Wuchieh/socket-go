package socket

import (
	"encoding/json"
)

type _req struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func handlerError(s *Socket, m *Member, err error) {
	if s == nil || m == nil {
		return
	}
	c := new(Context)
	c.reset()
	c.m = m
	c.s = s
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
		logf("no handler for event %s", req.Event)
		return
	}

	c := new(Context)
	c.reset()
	c.m = m
	c.s = s
	c.Event = req.Event
	c.handlers = handlers
	c.Data = req.Data
	c.Next()
}
