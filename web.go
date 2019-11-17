package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type WebAPIError struct {
	ErrorString string `json:"error"`
	ErrorCode   int    `json:"code"`
}

func NewWebAPIError(code int, errorString string) WebAPIError {
	return WebAPIError{
		ErrorString: errorString,
		ErrorCode:   code,
	}
}

func WebWrite(w http.ResponseWriter, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error in web write: %s\n", err)
		return WebWriteError(w, 500, "internal server error")
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bs)

	return err
}

func WebWriteError(w http.ResponseWriter, code int, errorString string) error {
	msg := NewWebAPIError(code, errorString)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	return WebWrite(w, msg)
}
