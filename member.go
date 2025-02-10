package socket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"slices"
	"sync"
)

type Member struct {
	c       *websocket.Conn
	s       *Socket
	atRooms []string
	mx      sync.Mutex
	id      uuid.UUID
	Values  sync.Map
}

func (m *Member) leaveRoom(room string) {
	if index := slices.Index(m.atRooms, room); index > -1 {
		m.atRooms = append(m.atRooms[:index], m.atRooms[index+1:]...)
	}
}

func (m *Member) joinRoom(room string) {
	if slices.Index(m.atRooms, room) == -1 {
		m.atRooms = append(m.atRooms, room)
	}
}

func NewMember(conn *websocket.Conn) *Member {
	return &Member{
		c:  conn,
		id: uuid.New(),
	}
}

func (m *Member) getRooms() []string {
	return m.atRooms
}

func (m *Member) GetRooms() []string {
	return m.getRooms()
}

func (m *Member) Close() error {
	if m.s != nil {
		m.s.CloseMember(m)
	}
	return m.c.Close()
}

func (m *Member) Listen() {
	defer func() {
		handlerOnDisconnect(m.s, m)
		_ = m.Close()
	}()

	handlerOnConnect(m.s, m)

	for {
		_, msg, err := m.c.ReadMessage()
		if err != nil {
			handlerError(m.s, m, err)
			return
		}
		handlerMessage(m.s, m, msg)
	}
}

func (m *Member) Emit(e string, data any) error {
	r := Response{
		Event: e,
		Data:  data,
	}
	b := r.GetByte()
	return m.WriteMessage(websocket.TextMessage, b)
}

func (m *Member) WritePreparedMessage(pm *websocket.PreparedMessage) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	return m.c.WritePreparedMessage(pm)
}

func (m *Member) WriteMessage(messageType int, data []byte) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	return m.c.WriteMessage(messageType, data)
}
