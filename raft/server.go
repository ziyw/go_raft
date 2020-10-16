package raft

import (
	"context"
	"fmt"
	"github.com/ziyw/go_raft/file"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	Name string
	Addr string
	Id   string
	Role string

	StopAll chan struct{}

	followers []*Server
	leader    *Server

	nextIndex  []int
	matchIndex []int

	commitIndex int
	lastApplied int

	termFile string
	voteFile string
	logFile  string
}

func NewServer(name, addr, id, role string) *Server {
	s := &Server{
		Name:        name,
		Addr:        addr,
		Id:          id,
		Role:        role,
		StopAll:     make(chan struct{}),
		commitIndex: 0,
		lastApplied: 0,
		termFile:    fmt.Sprintf("./temp/%s_term", addr),
		voteFile:    fmt.Sprintf("./temp/%s_vote", addr),
		logFile:     fmt.Sprintf("./temp/%s_log", addr),
	}
	s.SetCurrentTerm(1)
	s.SetVotedFor("")
	log := []*pb.Entry{&pb.Entry{Term: 0, Command: ""}}
	s.SetLog(log)
	return s
}

func (s *Server) Start(ctx context.Context, cancel context.CancelFunc) {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRaftServiceServer(grpcServer, s)

	log.Printf("Server %s start serving...", s.Addr)
	go grpcServer.Serve(lis)

	if s.Role == "l" {
		s.nextIndex = make([]int, len(s.followers))
		s.matchIndex = make([]int, len(s.followers))
		// TODO: add this back later
		go s.StartHeartbeat(ctx, cancel)
	}

	select {
	case <-ctx.Done():
		log.Printf("Server %s stopped", s.Addr)
		grpcServer.Stop()
		return
	}
}

// Raft interface impl
func (s *Server) Query(ctx context.Context, arg *pb.QueryArg) (*pb.QueryRes, error) {
	ctx, cancel := context.WithCancel(ctx)
	return s.HandleQuery(ctx, cancel, arg)
}

func (s *Server) RequestVote(ctx context.Context, arg *pb.VoteArg) (*pb.VoteRes, error) {
	return s.HandleRequestVote(ctx, arg)
}

func (s *Server) AppendEntries(ctx context.Context, arg *pb.AppendArg) (*pb.AppendRes, error) {
	return s.HandleAppendEntries(ctx, arg)
}

// Persist properties
func (s *Server) CurrentTerm() int {
	lines, err := file.ReadLines(s.termFile)
	if err != nil {
		return -1
	}
	t, err := strconv.ParseInt(strings.Trim(lines[0], "\n"), 10, 64)
	if err != nil {
		return -1
	}
	return int(t)
}

func (s *Server) SetCurrentTerm(t int) {
	f := s.termFile
	l := fmt.Sprintf("%d\n", t)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		file.AppendLines(f, []string{l})
		return
	}
	file.DeleteFile(f)
	file.AppendLines(f, []string{l})
}

func (s *Server) VotedFor() string {
	f := s.voteFile
	lines, err := file.ReadLines(f)
	if err != nil {
		return ""
	}
	return strings.Trim(lines[0], "\n ")
}

func (s *Server) SetVotedFor(v string) {
	f := s.voteFile
	l := fmt.Sprintf("%s\n", v)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		file.AppendLines(f, []string{l})
		return
	}
	file.DeleteFile(f)
	file.AppendLines(f, []string{l})
}

// TODO: Can be optimized
func (s *Server) Log() []*pb.Entry {
	f := s.logFile
	lines, err := file.ReadLines(f)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	entries := []*pb.Entry{}
	for _, l := range lines {
		p := strings.Split(l, ",")
		t, err := strconv.ParseInt(p[0], 10, 64)
		if err != nil {
			panic(err)
		}
		c := strings.Join(p[1:], ",")
		e := &pb.Entry{Term: t, Command: c}
		entries = append(entries, e)
	}
	return entries
}

func (s *Server) SetLog(entries []*pb.Entry) {
	lines := []string{}
	for _, e := range entries {
		l := fmt.Sprintf("%d,%s\n", e.Term, e.Command)
		lines = append(lines, l)
	}

	f := s.logFile
	file.DeleteFile(f)
	file.AppendLines(f, lines)
}

func (s *Server) AppendLog(e *pb.Entry) {
	l := fmt.Sprintf("%d,%s\n", e.Term, e.Command)
	file.AppendLines(s.logFile, []string{l})
}
