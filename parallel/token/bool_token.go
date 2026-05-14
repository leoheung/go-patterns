package token

import "sync"

type BoolToken struct {
	value bool
	mu    sync.Mutex
}

func NewBoolToken(value bool) *BoolToken {
	return &BoolToken{value: value}
}

func (bt *BoolToken) Get() bool {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	return bt.value
}

func (bt *BoolToken) Set(value bool) {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	bt.value = value
}
