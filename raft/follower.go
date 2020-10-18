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

func (s *Server) StartElectionTimeout(ctx context.Context, cancel context.CancelFunc) {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	timeout := func() time.Duration {
		return time.Duration((r.Int31n(150) + 150)) * time.Millisecond
	}

	ticker := time.NewTicker(timeout())
	defer ticker.Stop()

	for {
		log.Printf("Server %s: new election timeout round", s.Addr)
		select {
		case <-s.stopFollow:
			log.Printf("Server %s: following ended", s.Addr)
			cancel()
			return

		case <-s.stopElectionTimeout:
			ticker.Reset(timeout())
			log.Printf("Server %s: cancel this round of election timeout", s.Addr)
			continue
		case <-ctx.Done():
			log.Printf("Server %s: cancelled", s.Addr)
			return
		case <-ticker.C:
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
				s.stopFollow <- struct{}{}
				s.StartLeader(ctx, cancel)
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
