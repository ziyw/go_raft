package main

import (
	"context"
	_ "fmt"
	"github.com/ziyw/go_raft/raft"
	_ "os"
	"time"
)

func main() {
	// 	if len(os.Args) == 1 || os.Args[1] == "start" {
	// 		done := make(chan int)
	// 		raft.NewCluster("config")
	// 		<-done
	// 	}
	//
	// 	if len(os.Args) == 3 {
	// 		raft.SendRequest(os.Args[1], os.Args[2])
	// 	}
	// Build a single leader, send heartbeat to fake addr
	leader := raft.NewServer("Leader", "localhost:50001", "s1", "l", []string{"localhost:50002", "localhost:50003", "localhost:50004"})

	ctx, cancel := context.WithCancel(context.Background())
	leader.Start(ctx, cancel)

	leader.StartHeartbeat <- struct{}{}
	time.Sleep(1 * time.Second)
	leader.StopHeartbeat <- struct{}{}
	time.Sleep(1 * time.Second)

	leader.StartHeartbeat <- struct{}{}
	time.Sleep(1 * time.Second)
	leader.StopHeartbeat <- struct{}{}
	time.Sleep(1 * time.Second)

	leader.StartHeartbeat <- struct{}{}
	time.Sleep(1 * time.Second)
	leader.StopHeartbeat <- struct{}{}
	time.Sleep(1 * time.Second)
}
