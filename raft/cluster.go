package raft

import (
	_ "go_raft/pb"
	"log"
	"strings"
)

type Cluster struct {
	Servers []*Server
}

// Parse config file to create servers.
func (c *Cluster) Config(file string) []*Server {
	lines, err := ReadLines(file)
	if err != nil {
		log.Fatal(err)
	}

	s := []*Server{}
	for _, l := range lines {
		p := strings.Split(l, ",")
		s = append(s, NewServer(p[0], p[1], p[2], p[3]))
	}
	c.Servers = s
	return s
}

// Start all servers in cluster.
func (c Cluster) Start() {
	for _, s := range c.Servers {
		go s.Start()
	}
}
