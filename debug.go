package main

import (
	"fmt"
	"os"
	"time"
)

func Log(msg string) {
	f, err := os.OpenFile("air.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// Cannot log, but we can't stop the app for this.
		return
	}
	defer f.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(f, "[%s] %s\n", timestamp, msg)
}
