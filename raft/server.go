package raft

import (
	_ "context"
	_ "fmt"
	"github.com/ziyw/go_raft/pb"
	_ "golang.org/x/net/context"
	"google.golang.org/grpc"
	_ "log"
	"net"
	_ "os"
	_ "strconv"
	_ "strings"
)

type Server struct {
	Name string
	Addr string
	Id   string
	Role string

	followers []*Server
	leader    *Server

	nextIndex  []int
	matchIndex []int

	commitIndex int
	lastApplied int
}

// Persist properties
func (s *Server) CurrentTerm() int     { return 10 }
func (s *Server) SetCurrentTerm(t int) {}

func (s *Server) VotedFor() int     { return 0 }
func (s *Server) SetVotedFor(v int) {}

func (s *Server) Log() []*pb.Entry   { return nil }
func (s *Server) SetLog([]*pb.Entry) {}

// Common server methods

func NewServer(name, addr, id, role string) *Server {
	s := &Server{
		Name:        name,
		Addr:        addr,
		Id:          id,
		Role:        role,
		commitIndex: 0,
		lastApplied: 0,
	}
	s.SetCurrentTerm(1)
	s.SetVotedFor(-1)
	s.SetLog(nil)
	return s
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	_s := grpc.NewServer()
	// pb.RegisterRaftServiceServer(_s, s)
	if err := _s.Serve(lis); err != nil {
		return err
	}
	return nil
}
