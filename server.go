package main

import (
	"fmt"
	_ "golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
)

type Node interface {
	Start()
	IsLeader() bool
}

type Server struct {
	Name string
	Addr string

	// leader state
	nextIndex  []uint32
	matchIndex []uint32

	// common server state
	commitIndex uint32
	lastApplied uint32

	// persist states, TODO: need to write those to file
	currentTerm uint32
	votedFor    int // when int < 0, is nil
	log         []Entry
}

func (s *Server) Start() {
	fmt.Printf("Start server binding Name: %s Addr: %s\n", s.Name, s.Addr)

	lis, err := net.Listen("tcp", s.Addr)
	Check(err)

	_s := grpc.NewServer()
	RegisterRaftServiceServer(_s, s)
	err = _s.Serve(lis)
	Check(err)
}

// TODO
func (s Server) IsLeader() bool { return false }
