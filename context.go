package socket

import (
	"encoding/json"
	"math"
	"reflect"
)

const abortIndex int8 = math.MaxInt8 >> 1

type HandlerFunc func(c *Context)

type Context struct {
	handlers []HandlerFunc
	index    int8
	s        *Socket
	m        *Member
	values   map[string]interface{}

	Data  any
	Event string
}

func (c *Context) reset() {
	c.handlers = nil
	c.index = -1
	c.s = nil
	c.m = nil
	c.values = nil
	c.Data = nil
	c.Event = ""
}

func (c *Context) GetSocket() *Socket {
	return c.s
}

func (c *Context) GetMember() *Member {
	return c.m
}

func (c *Context) Set(key string, value interface{}) {
	if c.values == nil {
		c.values = make(map[string]interface{})
	}
	c.values[key] = value
}

func (c *Context) Get(key string) (val any, exists bool) {
	val, exists = c.values[key]
	return
}

// Join 加入房間
func (c *Context) Join(room string) {
	c.m.mx.Lock()
	defer c.m.mx.Unlock()

	JoinRoom(room, c.s, c.m)
}

// Leave 離開房間
func (c *Context) Leave(room string) {
	c.m.mx.Lock()
	defer c.m.mx.Unlock()

	LeaveRoom(room, c.s, c.m)
}

func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

func (c *Context) Abort() {
	c.index = abortIndex
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

// To 設置接收訊息的房間
func (c *Context) To(room string) *ContextTo {
	return &ContextTo{
		c:  c,
		to: []string{room},
	}
}

// Except 排除在該房間的用戶
func (c *Context) Except(room string) *ContextTo {
	return &ContextTo{
		c:      c,
		except: []string{room},
	}
}

// Emit 訊息只會傳給觸發者
func (c *Context) Emit(e string, data any) error {
	return c.m.Emit(e, data)
}

// Bind 簡單粗暴的當作 json 解析綁定
func (c *Context) Bind(obj any) error {
	objVal := reflect.ValueOf(obj)
	if objVal.Kind() != reflect.Ptr || objVal.IsNil() {
		return ErrBinData
	}

	b, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &obj)
}
