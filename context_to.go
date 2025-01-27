package socket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"slices"
	"sync"
)

type ContextTo struct {
	c *Context

	err error

	// 發送至哪些room
	to []string

	// 哪些room房間的人收不到訊息
	except []string
}

func (c *ContextTo) copy() *ContextTo {
	return &ContextTo{
		c:      c.c,
		to:     c.to,
		except: c.except,
	}
}

func (c *ContextTo) checkToExcept() bool {
	if len(c.to) == 0 {
		c.err = ErrToListEmpty
		return false
	}

	for _, s := range c.except {
		if slices.Contains(c.to, s) {
			c.err = ErrToExceptDuplicates
			return false
		}
	}

	return true
}

func (c *ContextTo) getExceptMembers() []*Member {
	m := make(map[uuid.UUID]*Member)

	for _, s := range c.except {
		// 先找到房間
		value, ok := c.c.s.rooms.Load(s)
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

	// 用 append 直接構建 members 切片
	members := make([]*Member, 0, len(m))
	for _, member := range m {
		members = append(members, member)
	}

	return members
}

func (c *ContextTo) getToMembers() []*Member {
	m := make(map[uuid.UUID]*Member)

	for _, s := range c.to {
		// 先找到房間
		value, ok := c.c.s.rooms.Load(s)
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

	// 用 append 直接構建 members 切片
	members := make([]*Member, 0, len(m))
	for _, member := range m {
		members = append(members, member)
	}

	return members
}

func (c *ContextTo) getMembers() []*Member {
	to := c.getToMembers()
	except := c.getExceptMembers()

	toMap := make(map[*Member]struct{})
	for _, member := range to {
		toMap[member] = struct{}{}
	}

	for _, member := range except {
		delete(toMap, member)
	}

	var result []*Member
	for member := range toMap {
		result = append(result, member)
	}

	return result
}

func (c *ContextTo) Set(key string, value interface{}) {
	c.c.Set(key, value)
}

func (c *ContextTo) Get(key string) (val any, exists bool) {
	val, exists = c.c.values[key]
	return
}

// To 設置接收訊息的房間
func (c *ContextTo) To(room string) *ContextTo {
	if slices.Contains(c.to, room) {
		return c
	}
	cp := c.copy()
	cp.to = append(cp.to, room)
	return cp
}

// Except 排除在該房間的用戶
func (c *ContextTo) Except(room string) *ContextTo {
	if slices.Contains(c.except, room) {
		return c
	}
	cp := c.copy()
	cp.except = append(cp.except, room)
	return cp
}

func (c *ContextTo) Emit(e string, data any) error {
	if c.err != nil {
		return c.err
	}

	r := Response{
		Event: e,
		Data:  data,
	}
	pm, err := websocket.NewPreparedMessage(websocket.TextMessage, r.GetByte())
	if err != nil {
		return err
	}

	m := c.getMembers()
	if len(m) == 0 {
		return ErrToListEmpty
	}

	wg := sync.WaitGroup{}
	wg.Add(len(m))
	for _, member := range m {
		go func() {
			defer wg.Done()
			err := member.WritePreparedMessage(pm)
			if err != nil {
				c.err = addEmitErr(c.err, EmitError{
					Member: member,
					Err:    err,
				})
			}
		}()
	}

	wg.Wait()
	return c.err
}
