package test

import (
	"testing"
	"time"

	"iop.inspur.com/common/loggerfactory"
)
var logger = loggerfactory.GetLogger()
func TestLogger(t *testing.T) {

	index := 1
	for {
		if index > 40 {
			break
		}
		index ++
		time.Sleep(time.Second * 2)
		logger.Info("info%s %d", "uuuuuuuuuuuuu", index)
		logger.Error("error%s", "uuuuuuuuuuuuu")
		logger.Debug("debug%s", "uuuuuuuuuuuuu")
		logger.Warn("warn%s", "uuuuuuuuuuuuu")
	}
	logger.Close()
}


func TestLogger2(t *testing.T) {
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
