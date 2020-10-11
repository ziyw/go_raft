package main

import (
	"github.com/ziyw/go_raft/raft"
	_ "log"
)

func main() {
	raft.NewCluster("config")
}
