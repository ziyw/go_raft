package main

import (
	_ "context"
	_ "fmt"
	"github.com/ziyw/go_raft/raft"
	"os"
)

func main() {
	if len(os.Args) == 1 || os.Args[1] == "start" {
		done := make(chan int)
		raft.NewCluster("config")
		<-done
	}

	if len(os.Args) == 3 {
		raft.SendRequest(os.Args[1], os.Args[2])
	}
}
