package session

import (
	"sync"
)

type RingBuffer struct {
	buf    []byte
	size   int
	r      int
	w      int
	length int
	mu     sync.Mutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buf:  make([]byte, size),
		size: size,
	}
}

func (rb *RingBuffer) Write(p []byte) (n int, err error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	n = len(p)
	if n == 0 {
		return 0, nil
	}

	if n > rb.size {
		p = p[n-rb.size:]
		rb.w = 0
		rb.r = 0
		rb.length = rb.size
		copy(rb.buf, p)
		return n, nil
	}

	free := rb.size - rb.length

	if n > free {
		overwrite := n - free
		rb.r = (rb.r + overwrite) % rb.size
		rb.length = rb.size
	} else {
		rb.length += n
	}

	firstChunk := rb.size - rb.w
	if n <= firstChunk {
		copy(rb.buf[rb.w:], p)
		rb.w = (rb.w + n) % rb.size
	} else {
		copy(rb.buf[rb.w:], p[:firstChunk])
		copy(rb.buf[0:], p[firstChunk:])
		rb.w = n - firstChunk
	}

	return len(p), nil
}

func (rb *RingBuffer) Bytes() []byte {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	out := make([]byte, rb.length)

	if rb.length == 0 {
		return out
	}

	firstChunk := rb.size - rb.r
	if rb.length <= firstChunk {
		copy(out, rb.buf[rb.r:rb.r+rb.length])
	} else {
		copy(out[:firstChunk], rb.buf[rb.r:])
		copy(out[firstChunk:], rb.buf[:rb.length-firstChunk])
	}

	return out
}
