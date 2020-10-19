package main

import (
	_ "context"
	_ "fmt"
	"github.com/ziyw/go_raft/raft"
	"os"
	"time"
)

func main() {
	if len(os.Args) == 3 {
		raft.SendRequest(os.Args[1], os.Args[2])
	}

	cluster := raft.NewCluster("config")
	cluster[0].StartHeartbeat <- struct{}{}
	for _, c := range cluster[1:] {
		c.TryStartTimeout()
	}

	time.Sleep(1 * time.Second)
	cluster[0].StopHeartbeat <- struct{}{}

	// Build a single leader, send heartbeat to fake addr
	// leader := raft.NewServer("Leader", "localhost:50001", "s1", "l", []string{"localhost:50002", "localhost:50003", "localhost:50004"})
	// ctx, cancel := context.WithCancel(context.Background())
	// leader.Start(ctx, cancel)

	// leader.StartTimeout <- struct{}{}
	// time.Sleep(1 * time.Second)

	// leader.StopTimeoutLogic()

	// leader.StartTimeout <- struct{}{}
	// time.Sleep(30 * time.Millisecond)
	// leader.HeardFromLeader <- struct{}{}
	// time.Sleep(1 * time.Second)
	//
	//	leader.StartHeartbeat <- struct{}{}
	//	time.Sleep(1 * time.Second)
	//	leader.StopHeartbeat <- struct{}{}
	//	time.Sleep(1 * time.Second)
	//
	//	leader.StartHeartbeat <- struct{}{}
	//	time.Sleep(1 * time.Second)
	//	leader.StopHeartbeat <- struct{}{}
	//	time.Sleep(1 * time.Second)
	//
	//	leader.StartHeartbeat <- struct{}{}
	//	time.Sleep(1 * time.Second)
	//	leader.StopHeartbeat <- struct{}{}
	//	time.Sleep(1 * time.Second)
	//
}
