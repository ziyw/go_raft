package raft

import (
	_ "fmt"
	_ "github.com/ziyw/go_raft/pb"
	_ "golang.org/x/net/context"
	_ "google.golang.org/grpc"
	_ "log"
	_ "net"
	_ "os"
	_ "strconv"
	_ "strings"
)

// Common server behavior

// TODO: 1 Create server config file to create all servers -> Long running files
// TODO: - make main() run every server and just keep running
// TODO: move sending request logic to another cli
// TODO: server behavior: election timeout
// TODO: leader behavior: check why there is one log entry losing

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
func CurrentTerm()    {}
func SetCurrentTerm() {}

func VotedFor()    {}
func SetVotedFor() {}

func Log()    {}
func SetLog() {}

// Common server methods

func NewServer(name, addr, id, role string) *Server { return nil }

func (s *Server) Start() error {
	return nil
}
