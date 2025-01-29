package socket

import (
	"slices"
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
	s.RoomJoin(room, m)

	if slices.Index(m.atRooms, room) == -1 {
		m.atRooms = append(m.atRooms, room)
	}
}

// LeaveRoom unsafe
func LeaveRoom(room string, s *Socket, m *Member) {
	s.RoomLeave(room, m)

	if index := slices.Index(m.atRooms, room); index > -1 {
		m.atRooms = append(m.atRooms[:index], m.atRooms[index+1:]...)
	}
}
