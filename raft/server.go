package raft

import (
	"context"
	"fmt"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	Name    string
	Addr    string
	Id      string
	Role    string
	Cluster []string // TODO: persist cluster info?

	nextIndex  []int
	matchIndex []int

	commitIndex int
	lastApplied int

	termFile string
	voteFile string
	logFile  string

	// Leader behavior
	StartHeartbeat   chan struct{}
	StopHeartbeat    chan struct{}
	HeartbeatRunning bool

	// Follower behavior
	StartTimeout    chan struct{}
	StopTimeout     chan struct{}
	HeardFromLeader chan struct{}
	TimeoutRunning  bool
}

func NewServer(name, addr, id, role string, cluster []string) *Server {
	s := &Server{
		Name:        name,
		Addr:        addr,
		Id:          id,
		Role:        role,
		Cluster:     cluster,
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

	s.StartHeartbeat = make(chan struct{})
	s.StopHeartbeat = make(chan struct{})
	s.HeartbeatRunning = false

	s.StartTimeout = make(chan struct{})
	s.StopTimeout = make(chan struct{})
	s.HeardFromLeader = make(chan struct{})
	s.TimeoutRunning = false
	return s
}

func (s *Server) Start(ctx context.Context, cancel context.CancelFunc) {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterRaftServiceServer(grpcServer, s)

	// Start grpc server
	log.Printf("Server %s start serving...\n", s.Addr)
	go grpcServer.Serve(lis)

	// start raft logic
	go s.RunRaft(ctx, cancel)
}

func (s *Server) RunRaft(ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("Server is done")
			return
		case <-s.StartHeartbeat:
			s.nextIndex = make([]int, len(s.Cluster))
			s.matchIndex = make([]int, len(s.Cluster))
			go s.SendingHeartbeat(ctx)
		case <-s.StartTimeout:
			go s.StartElectionTimeout(ctx)
		}
	}
}

// Raft interface impl
func (s *Server) Query(ctx context.Context, arg *pb.QueryArg) (*pb.QueryRes, error) {
	// 	return s.HandleQuery(ctx, cancel, arg)
	return nil, nil
}

func (s *Server) RequestVote(ctx context.Context, arg *pb.VoteArg) (*pb.VoteRes, error) {
	//	return s.HandleRequestVote(ctx, arg)
	return nil, nil
}

func (s *Server) AppendEntries(ctx context.Context, arg *pb.AppendArg) (*pb.AppendRes, error) {
	return s.HandleAppendEntries(ctx, arg)
}
