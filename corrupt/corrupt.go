package corrupt

import (
	"io"
	"time"
	"math/rand"
)

func New(f io.ReadWriter) io.ReadWriter {
	rw := &corruptingWrap{rw: f}
	rw.corrupt.nmsgMax = time.Now().Nanosecond() & 7
	return rw
}

func Wrap(f io.ReadWriter, _ string) io.ReadWriter {
	return New(f)
}

type corruptingWrap struct {
	rw      io.ReadWriter
	corrupt struct {
		tPrevW time.Time
		tPrevR time.Time
		n      int
		nmsg int
		nmsgMax int
	}
}

func (w *corruptingWrap) Read(buf []byte) (n int, err error) {
retry:
	n, err = w.rw.Read(buf)
	if err == nil && n > 0 {
		if corrupt.fakeRxActive() {
			n = corrupt.copyRx(buf)
			return
		}
		t := time.Now()
		if Δt := t.Sub(w.corrupt.tPrevR); w.corrupt.nmsg > w.corrupt.nmsgMax || w.corrupt.nmsg > 0 && Δt > 2000*time.Millisecond {
			if (w.corrupt.n & 3) == 1 {
				buf[n-1]++
			} else {
				n--
			}
			w.corrupt.nmsg = 0;
			w.corrupt.nmsgMax = 3+rand.Intn(17)
			w.corrupt.tPrevR = t
			if n == 0 {
				goto retry
			}
		}
	}
	w.corrupt.nmsg++
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

	n, err = w.rw.Write(buf)
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
