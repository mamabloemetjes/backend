package middleware

import "github.com/MonkyMars/gecho"

type Middleware struct {
	logger gecho.Logger
}

func NewMiddleware() *Middleware {
	return &Middleware{
		logger: *gecho.NewDefaultLogger(),
	}
}
