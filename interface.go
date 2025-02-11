package socket

type Emit interface {
	Emit(e string, d any) error
}

type To interface {
	To(room string) *ContextTo
}
