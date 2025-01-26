package socket

import "log"

func logf(format string, data ...any) {
	log.Printf(format, data...)
}
