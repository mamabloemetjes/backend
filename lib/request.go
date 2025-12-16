package lib

import (
	"encoding/json"
	"net/http"
)

func ExtractAndValidateBody[T any](r *http.Request) (*T, error) {
	var body T
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return &body, nil
}
