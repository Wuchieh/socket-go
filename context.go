package socket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math"
	"slices"
	"sync"
)

const abortIndex int8 = math.MaxInt8 >> 1

type HandlerFunc func(c *Context)

type Context struct {
	handlers []HandlerFunc
	index    int8
	s        *Socket
	m        *Member

	// 發送至哪些room
	to []string

	Data   any
	Event  string
	Values map[string]interface{}
}

func (c *Context) reset() {
	c.handlers = nil
	c.index = -1
	c.s = nil
	c.m = nil

	c.Event = ""
	c.Values = nil
}

func (c *Context) getMembers() []*Member {
	m := make(map[uuid.UUID]*Member)

	if len(c.to) == 0 {
		m[c.m.id] = c.m
	} else {
		for _, s := range c.to {
			// 先找到房間
			value, ok := c.s.rooms.Load(s)
			if !ok {
				continue
			}
			// 檢查物件類型
			room, ok := value.(*sync.Map)
			if !ok {
				continue
			}

			room.Range(func(key, value any) bool {
				id := key.(uuid.UUID)
				member := value.(*Member)
				m[id] = member
				return true
			})
		}
	}

	// 用 append 直接構建 members 切片
	members := make([]*Member, 0, len(m))
	for _, member := range m {
		members = append(members, member)
	}

	return members
}

func (c *Context) Set(key string, value interface{}) {
	if c.Values == nil {
		c.Values = make(map[string]interface{})
	}
	c.Values[key] = value
}

func (c *Context) Get(key string) (val any, exists bool) {
	val, exists = c.Values[key]
	return
}

// Join 加入房間
func (c *Context) Join(room string) {
	c.m.mx.Lock()
	defer c.m.mx.Unlock()

	c.s.RoomJoin(room, c.m)

	if slices.Index(c.m.atRooms, room) == -1 {
		c.m.atRooms = append(c.m.atRooms, room)
	}
}

// Leave 離開房間
func (c *Context) Leave(room string) {
	c.m.mx.Lock()
	defer c.m.mx.Unlock()

	c.s.RoomLeave(room, c.m)
	index := slices.Index(c.m.atRooms, room)
	if index > -1 {
		c.m.atRooms = append(c.m.atRooms[:index], c.m.atRooms[index+1:]...)
	}
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

func (c *Context) To(room string) *Context {
	if !slices.Contains(c.to, room) {
		c.to = append(c.to, room)
	}
	return c
}

func (c *Context) Emit(e string, data any) error {
	m := c.getMembers()
	if len(m) == 0 {
		return nil
	} else if len(m) == 1 {
		return m[0].Emit(e, data)
	} else {
		r := Response{
			Event: e,
			Data:  data,
		}
		pm, err := websocket.NewPreparedMessage(websocket.TextMessage, r.GetByte())
		if err != nil {
			return err
		}
		for _, member := range m {
			go func() {
				err := member.WritePreparedMessage(pm)
				if err != nil {
					logf("WritePreparedMessage Error: %s", err.Error())
				}
			}()
		}
	}
	return nil
}
