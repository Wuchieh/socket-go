package socket

import "sync"

type BroadcastTo struct {
	s      *Socket
	c      *Context
	except *Member
}

func (b *BroadcastTo) Get(key string) (val any, exists bool) {
	if b == nil || b.c == nil {
		return nil, false
	}

	return b.c.Get(key)
}

func (b *BroadcastTo) Emit(e string, d any) error {
	if b == nil || b.s == nil {
		return ErrSocketNil
	}

	var err error
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 預先計算成員數量
	memberCount := 0
	b.s.members.Range(func(k, v interface{}) bool {
		memberCount++
		return true
	})

	wg.Add(memberCount)

	b.s.members.Range(func(k, v interface{}) bool {
		member := v.(*Member)
		if member == b.except {
			wg.Done() // 跳過此成員時也要減少計數器
			return true
		}

		go func(m *Member) {
			defer wg.Done()

			_err := m.Emit(e, d)
			if _err != nil {
				mu.Lock()
				err = addEmitErr(err, EmitError{
					Member: m,
					Err:    _err,
				})
				mu.Unlock()
			}
		}(member)

		return true
	})

	wg.Wait()

	return err
}
