package socket

import (
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
