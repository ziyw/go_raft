package raft

import (
	"context"
	_ "fmt"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"time"
)

func (s *Server) StartTimeout(ctx context.Context, cancel context.CancelFunc) {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	for {
		electionTimeout := time.Duration((r.Int31n(150) + 150)) * time.Millisecond
		select {
		case <-s.StopFollow:
			return
		case <-ctx.Done():
			log.Printf("Server %s: cancelled", s.Addr)
			return
		case <-time.After(electionTimeout):
			log.Printf("Follower %s waiting timeout, start voting logic", s.Addr)
			s.SetCurrentTerm(s.CurrentTerm() + 1)
			go s.StartVoteRequest(ctx, cancel)
			continue
		}
	}
}

func (s *Server) StartVoteRequest(ctx context.Context, cancel context.CancelFunc) {
	done := make(chan *pb.VoteRes)
	for _, addr := range s.Cluster {
		go s.SendVoteRequest(ctx, cancel, done, addr)
	}

	count := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("Server %s: cancelled\n", s.Addr)
			return
		case r := <-done:
			if r.Term > int64(s.CurrentTerm()) {
				cancel()
			}
			if r.VoteGranted {
				count++
			}
			if count >= len(s.Cluster)/2+1 {
				log.Printf("Server %s: change to leader now", s.Addr)
				s.StopFollow <- struct{}{}
				s.StartLead <- struct{}{}
				return
			}
		}
	}

}

func (s *Server) SendVoteRequest(ctx context.Context, cancel context.CancelFunc, done chan *pb.VoteRes, addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
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
	done <- r
}
