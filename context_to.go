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
	to := c.getToMembers() // 已去重

	// 邊界條件：若 except 為空，直接返回結果
	if len(c.except) == 0 {
		return to
	}

	// 預處理例外集合
	exceptSet := make(map[string]struct{}, len(c.except))
	for _, s := range c.except {
		exceptSet[s] = struct{}{}
	}

	// 預分配結果切片容量 (假設大部分情況需要過濾)
	result := make([]*Member, 0, len(to))

	// 單次線性掃描
	for _, member := range to {
		if !hasAnyCommonRoom(member.atRooms, exceptSet) {
			result = append(result, member)
		}
	}

	if len(result) == 0 {
		return nil // 保持與原碼相同行為
	}
	return result
}

// 檢查房間交集 (保持不變)
func hasAnyCommonRoom(rooms []string, exceptSet map[string]struct{}) bool {
	for _, room := range rooms {
		if _, exists := exceptSet[room]; exists {
			return true
		}
	}
	return false
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
