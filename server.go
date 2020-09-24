package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
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

	group []*Server
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

// Persist current term
func (s *Server) SetCurrentTerm(term int64) {
	v := fmt.Sprintf("%d", term)
	if err := WriteLine(s.Addr+"CurrentTerm", v); err != nil {
		log.Println("Read current term failed")
	}
}

func (s *Server) CurrentTerm() int64 {
	line, err := ReadLines(s.Addr + "CurrentTerm")
	if err != nil {
		log.Println("Persist Current Term Failed")
	}
	result := strings.Trim(line[0], "\n")
	term, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return int64(term)
}

func (s *Server) CheckCurrentTerm() bool {
	if _, err := os.Stat(s.Addr + "CurrentTerm"); err != nil {
		return false
	}
	return true
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

// Query is receive normal query from normal client.
func (s *Server) Query(ctx context.Context, arg *QueryArg) (*QueryRes, error) {
	if s.State == Leader {
		r := fmt.Sprintf("%s says Hi", s.Name)
		return &QueryRes{Success: true, Reply: r}, nil
	}

	l := s.GetVotedFor()
	// no known leader in the group
	if l == -1 {
		for i, v := range s.group {
			if v.State == Leader {
				l = i
			}
		}
	}
	if l == -1 {
		// TODO: replace this with trigger voteAction
		log.Fatal("No Valid Leader")
	}

	curLeader := s.group[l]
	conn, err := grpc.Dial(curLeader.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := NewRaftServiceClient(conn)
	return c.Query(context.Background(), arg)
}

func (s Server) GetVotedFor() int {
	return int(s.votedFor)
}

func (s *Server) GetLeader() (*Server, error) {
	var l *Server = nil
	for _, v := range s.group {
		if v.State == Leader {
			if l == nil {
				l = v
			} else {
				return nil, fmt.Errorf("More than one leader in the group")
			}
		}
	}
	if l == nil {
		return nil, fmt.Errorf("No leader in the group")
	}
	return l, nil
}
