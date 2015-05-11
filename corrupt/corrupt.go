package corrupt

import (
	"io"
	"time"
)

type corruptingWrap struct {
	io.ReadWriteCloser
	corrupt struct {
		tPrevW time.Time
		tPrevR time.Time
		n      int
	}
}

func (w *corruptingWrap) Read(buf []byte) (n int, err error) {
retry:
	n, err = w.ReadWriteCloser.Read(buf)
	if err == nil && n > 0 {
		if corrupt.fakeRxActive() {
			n = corrupt.copyRx(buf)
			return
		}
		t := time.Now()
		if Δt := t.Sub(w.corrupt.tPrevR); Δt > 2000*time.Millisecond {
			if (w.corrupt.n & 3) == 1 {
				buf[n-1]++
			} else {
				n--
			}
			w.corrupt.tPrevR = t
			if n == 0 {
				goto retry
			}
		}
	}
	w.corrupt.n++
	return
}
func (w *corruptingWrap) Write(buf []byte) (n int, err error) {
	var dn int

	n = len(buf)
	if n > 2 && w.corrupt.n%100 == 0 {
		if (w.corrupt.n & 1) == 1 {
			buf[n-1]++
		} else {
			buf = buf[:n-1]
			dn = 1
		}
	}
	w.corrupt.n++

	n, err = w.ReadWriteCloser.Write(buf)
	n += dn
	return
}

var corrupt corruptionState

type corruptionState struct {
	FakeRx []string
	i      int
}

func (c *corruptionState) fakeRxActive() bool {
	return len(c.FakeRx) != 0
}

func (c *corruptionState) copyRx(buf []byte) (n int) {
	if c.i >= len(c.FakeRx) {
		c.i = 0
	}
	s := c.FakeRx[c.i]
	c.i++
	copy(buf, s)
	return len(s)
}
