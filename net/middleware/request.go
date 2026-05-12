package middleware

import (
	"fmt"
	"net/http"

	"github.com/leoheung/go-patterns/net"
	"github.com/leoheung/go-patterns/utils"
)

type MiddlewareFn func(http.Handler) http.Handler
type ProcessReqFn func(*http.Request) error
type ProcessResFn func(data []byte, headers http.Header, httpCode int) error

func BuildRequestLevelMiddleware(processReq ProcessReqFn, maxReqbodyMB int) (MiddlewareFn, error) {
	if maxReqbodyMB < 0 {
		return nil, fmt.Errorf("maxReqbodyMB < 0")
	}

	if processReq == nil {
		return nil, fmt.Errorf("processReq is nil")
	}

	var ret MiddlewareFn

	ret = func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. get request info
			reqCopy, err := net.DeepCopyRequest(r, maxReqbodyMB)
			if err != nil {
				utils.LogMessage(fmt.Sprintf("failed to deep copy request: %s", err.Error()))
			} else {
				go func() {
					err = processReq(reqCopy)
					if err != nil {
						utils.LogMessage(fmt.Sprintf("failed to process request: %s", err.Error()))
					}
				}()
			}
			// 2. get response info? todo

			next.ServeHTTP(w, r)
		})
	}

	return ret, nil
}
