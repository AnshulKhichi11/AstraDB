package engine

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type objectIDGen struct {
	mu      sync.Mutex
	rand5   [5]byte
	counter uint32
}

func newObjectIDGen() *objectIDGen {
	g := &objectIDGen{}
	_, _ = rand.Read(g.rand5[:])
	g.counter = uint32(time.Now().UnixNano()) & 0xFFFFFF
	return g
}

// 12 bytes: 4 bytes timestamp + 5 bytes random + 3 bytes counter
func (g *objectIDGen) NewHex() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	ts := uint32(time.Now().Unix())

	var b [12]byte
	b[0] = byte(ts >> 24)
	b[1] = byte(ts >> 16)
	b[2] = byte(ts >> 8)
	b[3] = byte(ts)

	copy(b[4:9], g.rand5[:])

	g.counter = (g.counter + 1) & 0xFFFFFF
	b[9] = byte(g.counter >> 16)
	b[10] = byte(g.counter >> 8)
	b[11] = byte(g.counter)

	out := make([]byte, 24)
	hex.Encode(out, b[:])
	return string(out)
}
