package net

import (
	"encoding/json"
	"net/http"
)

// Bind decodes the JSON request body into T. It is a pure decode — no
// response is written; the handler decides how to respond on error
// (typically ReturnErrorResponse with 400). Field validation stays
// hand-written in the handler (convention: no third-party validator
// dependency).
//
//	func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
//	    req, err := net.Bind[CreateOrderRequest](r)
//	    if err != nil {
//	        net.ReturnErrorResponse(w, http.StatusBadRequest, "invalid request body")
//	        return
//	    }
//	    // manual checks follow, e.g.
//	    // if req.UserID == "" { net.ReturnErrorResponse(w, http.StatusBadRequest, "userId is required"); return }
//	}
func Bind[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, err
	}
	return v, nil
}
