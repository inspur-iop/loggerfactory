package test

import (
	"testing"
	"time"

)
func TestLogger3(t *testing.T) {

	index := 1
	for {
		if index > 40 {
			break
		}
		index ++
		time.Sleep(time.Second * 2)
		logger.Info("info %s %d", "uuuuuuuuuuuuu", index)
		logger.Error("%d error %s ", index, "uuuuuuuuuuuuu")
		logger.Debug("debug%s", "uuuuuuuuuuuuu")
		logger.Warn("warn%s", "uuuuuuuuuuuuu")
	}
	logger.Close()
}


func TestLogger4(t *testing.T) {
	index := 1
	for {
		if index > 20 {
			break
		}
		index ++
		time.Sleep(time.Second * 2)

		logger.Info("info%s %d", "aaaaaaaaaa", index)
		logger.Error("error%s", "aaaaaaaaaaa")
		logger.Debug("debug%s", "aaaaaaaaaaaaaa")
		logger.Warn("warn%s", "aaaaaaaaaaa")
	}
	logger.Close()
}
