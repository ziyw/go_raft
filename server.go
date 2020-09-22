package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	Leader    = iota
	Follower  = iota
	Candidate = iota
)

type Server struct {
	Name  string
	Addr  string
	Id    int
	State int

	group []Server
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

func NewServer(name string, addr string, id int) *Server {
	// TODO: replace log with log read from file
	testLog := []*Entry{
		&Entry{Term: 1, Command: "Hello"},
		&Entry{Term: 1, Command: "World"},
	}
	return &Server{
		Name:  name,
		Addr:  addr,
		Id:    id,
		State: Follower,
		// TODO: three of them should be replaced by persist read
		currentTerm: 1, // TODO: replace with read term file
		votedFor:    -1,
		log:         testLog,
		commitIndex: 0,
		lastApplied: 0,
	}

}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	_s := grpc.NewServer()
	RegisterRaftServiceServer(_s, s)
	if err := _s.Serve(lis); err != nil {
		return err
	}
	return nil
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

// Query from normal clients.
func (s *Server) Query(ctx context.Context, arg *QueryArg) (*QueryRes, error) {
	r := fmt.Sprintf("%s says Hi", s.Name)
	return &QueryRes{Success: true, Reply: r}, nil
}
