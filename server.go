package main

import (
	"fmt"
	_ "golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	_ "os"
	_ "strconv"
	_ "strings"
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

	leaderId int
	group    []*Server
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

// TODO: not consider the restart situation
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

// TODO: Add Reboot function

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

// Persist current term
func (s *Server) CurrentTerm() (int, error) {
	return ReadInt(s.Addr + "CurrentTerm")
}

func (s *Server) SetCurrentTerm(term int) {
	if err := SaveInt(s.Addr+"CurrentTerm", term); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) CheckCurrentTerm() bool {
	return PeekFile(s.Addr + "CurrentTerm")
}

// Persist votedFor
func (s *Server) VotedFor() (int, error) {
	if !s.CheckVotedFor() {
		return -1, fmt.Errorf("Not voting yet")
	}
	return ReadInt(s.Addr + "VotedFor")
}

func (s *Server) SetVotedFor(term int) {
	if err := SaveInt(s.Addr+"VotedFor", term); err != nil {
		log.Fatal(err)
	}
}

func (s *Server) CheckVotedFor() bool {
	return PeekFile(s.Addr + "VotedFor")
}

// Persist Logs
func (s *Server) SaveEntry(entry *Entry) error {
	return SaveEntry(s.Addr+"Log", entry)
}

func (s *Server) Log() ([]*Entry, error) {
	return ReadEntries(s.Addr + "Log")
}
