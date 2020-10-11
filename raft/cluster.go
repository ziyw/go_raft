package raft

import (
	"github.com/ziyw/go_raft/file"
	"log"
	"strings"
)

type Cluster struct {
	Servers []*Server
}

func NewCluster(cfg string) (*Cluster, error) {
	lines, err := file.ReadLines(cfg)
	if err != nil {
		return nil, err
	}

	s := []*Server{}
	for _, l := range lines {
		cur := strings.Trim(l, "\n")
		if len(cur) == 0 {
			continue
		}
		arg := strings.Split(cur, ",")
		log.Printf("Current line is %v", arg)
		s = append(s, NewServer(arg[0], arg[1], arg[2], arg[3]))
	}
	return &Cluster{Servers: s}, nil
}

// Start all servers in cluster.
func (c Cluster) Start() {
	for _, s := range c.Servers {
		go s.Start()
	}
}
