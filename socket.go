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

	onError []HandlerFunc

	handlers map[string][]HandlerFunc

	otherHandler []HandlerFunc

	onConnect []HandlerFunc

	onDisconnect []HandlerFunc
}

func NewSocket() *Socket {
	return New()
}

func New() *Socket {
	return &Socket{}
}

func Default() *Socket {
	s := New()
	s.SetGrader(websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})

	return s
}

func (s *Socket) OnError(_func ...HandlerFunc) {
	if len(_func) >= int(abortIndex) {
		panic("too many handlers")
	}

	s.onError = _func
}

// On
//
//	監聽用戶傳入事件
func (s *Socket) On(e string, _func ...HandlerFunc) {
	if s.handlers == nil {
		s.handlers = make(map[string][]HandlerFunc)
	}

	checkHandlerFunc(_func...)

	if _, ok := s.handlers[e]; ok {
		panic("duplicate handler")
	}

	s.handlers[e] = _func
}

func (s *Socket) OnConnect(_func ...HandlerFunc) {
	checkHandlerFunc(_func...)

	s.onConnect = _func
}

func (s *Socket) OnDisconnect(_func ...HandlerFunc) {
	checkHandlerFunc(_func...)

	s.onDisconnect = _func
}

func (s *Socket) OnOtherEvent(_func ...HandlerFunc) {
	checkHandlerFunc(_func...)

	s.otherHandler = _func
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

func (s *Socket) roomAddMember(room string, m *Member) {
	// 使用 LoadOrStore 同時初始化並獲取 roomMap
	value, _ := s.rooms.LoadOrStore(room, &sync.Map{})
	roomMap, _ := value.(*sync.Map)

	// 儲存 member 到 roomMap
	roomMap.Store(m.id, m)
}

func (s *Socket) roomRemoveMember(room string, m *Member) {
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

func (s *Socket) RoomJoin(room string, m *Member) {
	JoinRoom(room, s, m)
}

func (s *Socket) RoomLeave(room string, m *Member) {
	LeaveRoom(room, s, m)
}

func (s *Socket) Emit(event string, data any) error {
	m := GetMembers(&s.members)
	return Broadcast(m, event, data)
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
