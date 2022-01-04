package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"

	"github.com/triptych-labs/anchor-escrow/v2/src/staking"
)

func main_rpc() {

	client, err := rpc.DialHTTP("tcp", "0.0.0.0:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	// Synchronous call
	args := staking.Subscription{Test: fmt.Sprint(time.Now().Unix())}
	var reply int
	err = client.Call("Subscriptions.SubscribeToEvent", args, &reply)
	if err != nil {
		log.Fatal("arith error:", err)
	}

}
