package utils

import "github.com/lwch/logging"

func Recover(key string) {
	if err := recover(); err != nil {
		logging.Error("%s: %v", key, err)
	}
}
