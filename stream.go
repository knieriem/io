package stream

import (
	"io"
	"sync"
)

func InheritCloser(dest io.ReadWriter, src io.Closer) io.ReadWriteCloser {
	if i, ok := dest.(io.ReadWriteCloser); ok {
		if i == src {
			return src.(io.ReadWriteCloser)
		}
	}
	return &struct {
		io.ReadWriter
		io.Closer
	}{dest, src}
}

type WrapFunc func(f io.ReadWriter, connID string) io.ReadWriter

type Wrapper struct {
	mu sync.Mutex
	f  WrapFunc
}

func (w *Wrapper) Set(f WrapFunc) {
	w.mu.Lock()
	w.f = f
	w.mu.Unlock()
}

func (w *Wrapper) Wrap(rw io.ReadWriter, connID string) io.ReadWriter {
	w.mu.Lock()
	if w.f != nil {
		rw = w.f(rw, connID)
	}
	w.mu.Unlock()
	return rw
}
