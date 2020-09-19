package main

import (
	"fmt"
	_ "golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
	"strings"
)

type Server struct {
	Name string
	Addr string

	IsLeader bool

	// persist states, TODO: need to write those to file
	// TODO: current version, use in memory version
	currentTerm int64
	votedFor    int64 // when int < 0, is nil
	log         []*Entry

	// leader state
	nextIndex  []int
	matchIndex []int

	// common server state
	commitIndex int
	lastApplied int
}

func (s *Server) Start(success chan string) {
	fmt.Printf("Start server binding Name: %s Addr: %s\n", s.Name, s.Addr)

	lis, err := net.Listen("tcp", s.Addr)
	Check(err)

	_s := grpc.NewServer()
	RegisterRaftServiceServer(_s, s)
	success <- s.Name
	err = _s.Serve(lis)
	Check(err)
}

func (s *Server) SetCommitIndex(term int64) {
	if err := WriteLine(s.Name+"CommitIndex", string(term)); err != nil {
		log.Fatal(err)
	}
	s.currentTerm = term
}

func (s *Server) CommitIndex() int64 {
	line, err := ReadLines(s.Name + "CommitIndex")
	if err != nil {
		log.Fatal(err)
		return 0
	}

	result := strings.Trim(line[0], "\n")
	if term, err := strconv.ParseInt(result, 10, 64); err != nil {
		return 0
	} else {
		return int64(term)
	}
}
