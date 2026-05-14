package token

import (
	"fmt"
	"sync"
)

type StaticTokens struct {
	numTokens int
	mu        sync.Mutex
}

func NewStaticTokens(numTokens int) (*StaticTokens, error) {
	if numTokens <= 0 {
		return nil, fmt.Errorf("numTokens <= 0")
	}

	return &StaticTokens{
		numTokens: numTokens,
		mu:        sync.Mutex{},
	}, nil
}

func (st *StaticTokens) GrantNextToken() bool {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.numTokens <= 0 {
		return false
	}

	st.numTokens -= 1
	return true
}
