package queue

import (
	"fmt"
	"log"
	"time"
)

func QueueLogger(name string) *log.Logger {
	return log.New(log.Writer(), fmt.Sprintf("Queue - [%s] "+time.Now().Format("2006-01-02 15:04:05")+" ", name), log.LstdFlags)
}
