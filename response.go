package socket

import (
	"encoding/json"
	"sync"
)

const (
	EventError = "error"
)

type Response struct {
	Event string `json:"event"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`

	b  []byte
	mx sync.Mutex
}

func (r *Response) Copy() *Response {
	return &Response{Event: r.Event, Data: r.Data, Error: r.Error}
}

func ErrorResponse(err error) *Response {
	var m string
	if err != nil {
		m = err.Error()
	}

	return &Response{
		Event: EventError,
		Error: m,
	}
}

// GetByte 不會多次解析
func (r *Response) GetByte() []byte {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.b != nil {
		return r.b
	}

	r.b = r.ToByte()
	return r.b
}

func (r *Response) ToByte() []byte {
	c := r.Copy()
	b, err := json.Marshal(c)
	if err != nil {
		c.Error = err.Error()
		c.Data = nil
		b, _ = json.Marshal(c)
	}
	return b
}
