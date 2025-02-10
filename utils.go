package socket

import (
	"github.com/gorilla/websocket"
	"sync"
)

func MapLength(m *sync.Map) int {
	length := 0
	m.Range(func(key, value interface{}) bool {
		length++
		return true
	})
	return length
}

// JoinRoom unsafe
func JoinRoom(room string, s *Socket, m *Member) {
	s.roomAddMember(room, m)

	m.joinRoom(room)
}

// LeaveRoom unsafe
func LeaveRoom(room string, s *Socket, m *Member) {
	s.roomRemoveMember(room, m)

	m.leaveRoom(room)
}

func GetMembers(m *sync.Map) []*Member {
	if m == nil {
		return nil
	}

	var members []*Member

	m.Range(func(key, value interface{}) bool {
		member := value.(*Member)
		members = append(members, member)
		return true
	})

	return members
}

func Broadcast(m []*Member, event string, data any) error {
	res := Response{
		Event: event,
		Data:  data,
	}

	preparedMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, res.GetByte())
	if err != nil {
		return err
	}

	var eErr error
	var wg sync.WaitGroup

	wg.Add(len(m))
	for _, member := range m {
		go func(member *Member) {
			if err := member.c.WritePreparedMessage(preparedMsg); err != nil {
				eErr = addEmitErr(eErr, EmitError{
					Member: member,
					Err:    err,
				})
			}
		}(member)
	}

	wg.Wait()
	return eErr
}
