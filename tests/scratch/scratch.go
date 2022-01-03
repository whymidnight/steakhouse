package main

import (
	"fmt"
	"log"
	"time"

	"go.uber.org/atomic"
)

var a = make(map[string]atomic.Int64)

func main() {
	i := a["a"]
	fmt.Println(i.Load())
	i.Inc()
	fmt.Println(i.Load())
	fmt.Println(5 / float64(2))

	now := time.Now().UTC().Unix()
	endDate := now + (60 * 5) // in 5 minutes
	duration := endDate - now
	log.Println("Now:", now, "End Date:", endDate, "Duration:", duration)

	EVERY := int64(5)
	epochs := (int(duration / EVERY))
	for i := range make([]interface{}, epochs) {
		log.Println("sleeping for EVERY", EVERY, fmt.Sprint(i+1, "/", epochs))
		time.Sleep(time.Duration(EVERY) * time.Second)
	}
	rem := (int(endDate % EVERY))
	if rem != 0 {
		log.Println("Sleeping for REM", rem)
		time.Sleep(time.Duration(rem) * time.Millisecond)
	}
}

