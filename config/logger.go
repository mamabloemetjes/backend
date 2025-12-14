package config

import (
	"github.com/MonkyMars/gecho"
)

var logger gecho.Logger

func InitializeLogger() *gecho.Logger {
	logger = *gecho.NewDefaultLogger()
	return &logger
}

func GetLogger() *gecho.Logger {
	return &logger
}
