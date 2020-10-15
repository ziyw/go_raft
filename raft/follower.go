package raft

import (
	"context"
	_ "fmt"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	"time"
)

func (s *Server) StartListenHeartbeat(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Printf("Server %s stop listening ", s.Addr)
		return
	case <-time.After(time.Second):
		log.Printf("Server %s listen timeout, start voting now", s.Addr)
		s.StartVoteRequest(ctx)
		return
	}
}

func (s *Server) StartVoteRequest(ctx context.Context) {
	done := make(chan bool)
	go s.SendVoteRequest(ctx, done, s.leader)
	for _, f := range s.followers {
		go s.SendVoteRequest(ctx, done, f)
	}
}

func (s *Server) SendVoteRequest(ctx context.Context, done chan bool, f *Server) {
	conn, err := grpc.Dial(f.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	term := s.CurrentTerm()
	log := s.Log()
	arg := &pb.VoteArg{
		Term:         int64(term),
		CandidateId:  s.Id,
		LastLogIndex: int64(len(log) - 1),
		LastLogTerm:  int64(term),
	}

	client := pb.NewRaftServiceClient(conn)
	r, err := client.RequestVote(ctx, arg)
	done <- r.VoteGranted
}
