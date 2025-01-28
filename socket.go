package socket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type Socket struct {
	upgrader websocket.Upgrader

	// members	map[uuid.UUID]*Member
	members sync.Map

	// rooms	map[string]map[uuid.UUID]*Member
	rooms sync.Map

	handlers map[string][]HandlerFunc
}

func NewSocket() *Socket {
	return &Socket{}
}

// On
//
//	監聽用戶傳入事件
func (s *Socket) On(e string, _func ...HandlerFunc) {
	if s.handlers == nil {
		s.handlers = make(map[string][]HandlerFunc)
	}

	if len(_func) >= int(abortIndex) {
		panic("too many handlers")
	}

	if _, ok := s.handlers[e]; ok {
		panic("duplicate handler")
	}

	s.handlers[e] = _func
}

// CloseMember
//
//	關閉 member 連線
func (s *Socket) CloseMember(m *Member) {
	// 讓 member 離開所有房間
	for _, room := range m.atRooms {
		s.RoomLeave(room, m)
	}
	// 將 member 從 members 中移除
	s.members.Delete(m.id)
}

// SetGrader
//
//	設置 websocket.Upgrader
func (s *Socket) SetGrader(upgrader websocket.Upgrader) {
	s.upgrader = upgrader
}

// GetMember 取得 Member
func (s *Socket) GetMember(id uuid.UUID) (*Member, bool) {
	value, ok := s.members.Load(id)
	if ok {
		return value.(*Member), true
	}
	return nil, false
}

// CreateMember
//
//	創建 member
func (s *Socket) CreateMember(w http.ResponseWriter, r *http.Request) (*Member, error) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	member := NewMember(conn)
	member.s = s

	return member, nil
}

func (s *Socket) HandlerM(m *Member) {
	defer func(member *Member) {
		_ = member.Close()
	}(m)

	s.members.Store(m.id, m)
	m.Listen()
}

// Handler
//
//	簡單的 http HandleFunc
//	兼容 http 標準庫
func (s *Socket) Handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_ = r.Body.Close()
	}()

	startTime := time.Now()
	defer func() {
		logf(time.Now().Sub(startTime).String())
	}()

	member, err := s.CreateMember(w, r)
	if err != nil {
		logf(err.Error())
		return
	}

	s.HandlerM(member)

}

func (s *Socket) RoomJoin(room string, m *Member) {
	// 使用 LoadOrStore 同時初始化並獲取 roomMap
	value, _ := s.rooms.LoadOrStore(room, &sync.Map{})
	roomMap, _ := value.(*sync.Map)

	// 儲存 member 到 roomMap
	roomMap.Store(m.id, m)
}

func (s *Socket) RoomLeave(room string, m *Member) {
	// 從 rooms 中加載 roomMap
	value, ok := s.rooms.Load(room)
	if !ok {
		return
	}

	roomMap, ok := value.(*sync.Map)
	if ok {
		roomMap.Delete(m.id)
	}
}

func (s *Socket) Emit(event string, data any) error {
	// Prepare the message
	res := Response{
		Event: event,
		Data:  data,
	}

	preparedMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, res.GetByte())
	if err != nil {
		return err
	}

	var eErr error

	// Send the prepared message to all members
	s.members.Range(func(_, value interface{}) bool {
		if member, ok := value.(*Member); ok {
			go func(m *Member) {
				if err := m.WritePreparedMessage(preparedMsg); err != nil {
					eErr = addEmitErr(eErr, EmitError{
						Member: m,
						Err:    err,
					})
				}
			}(member)
		}
		return true
	})

	return eErr
}

func (s *Socket) To(room string) *ContextTo {
	c := new(Context)
	c.reset()
	c.s = s
	return c.To(room)
}

func (s *Socket) Except(room string) *ContextTo {
	c := new(Context)
	c.reset()
	c.s = s
	return c.Except(room)
}
