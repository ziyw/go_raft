package raft

import (
	"context"
	"fmt"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"time"
)

// Timeout logic.
func (s *Server) TryStartTimeout() {
	if s.TimeoutRunning {
		return
	}
	s.StartTimeout <- struct{}{}
}

func (s *Server) TryStopTimeout() {
	if !s.TimeoutRunning {
		return
	}
	s.StopTimeout <- struct{}{}
}

func (s *Server) StartElectionTimeout(ctx context.Context) {
	s.TimeoutRunning = true

	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	newTimeout := func() time.Duration {
		return time.Duration((r.Int31n(150) + 150)) * time.Millisecond
	}
	ticker := time.NewTicker(newTimeout())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.TimeoutRunning = false
			log.Printf("Server %s: cancelled", s.Addr)
			return
		case <-s.StopTimeout:
			s.TimeoutRunning = false
			log.Printf("Server %s: timeout logic stopped", s.Addr)
			return
		case <-s.HeardFromLeader:
			log.Printf("Server %s: Heard from leader, cancel timeout ", s.Addr)
			ticker.Reset(newTimeout())
			continue
		case <-ticker.C:
			s.TimeoutRunning = false
			log.Printf("Follower %s waiting timeout, start voting logic", s.Addr)
			s.SetCurrentTerm(s.CurrentTerm() + 1)
			log.Println("Start voting logic")
			return
			// go s.StartVoteRequest(ctx, cancel)
			// continue
		}
	}
}

// AppendEntries logic
func (s *Server) HandleAppendEntries(ctx context.Context, arg *pb.AppendArg) (*pb.AppendRes, error) {

	s.HeardFromLeader <- struct{}{}

	term := int64(s.CurrentTerm())
	if term < 0 {
		err := fmt.Errorf("Invalid term number")
		return nil, err
	}

	res := &pb.AppendRes{Term: term, Success: false}

	if arg.Term < term {
		log.Println("Follower should start voting logic")
		return res, fmt.Errorf("Leader out-of-date")
	}

	if arg.Term > term {
		s.SetCurrentTerm(int(arg.Term))
	}

	// Empty Heartbeat msg
	if arg.LeaderId == "" {
		res.Success = true
		return res, nil
	}

	logs := s.Log()
	if logs == nil {
		return nil, fmt.Errorf("Error getting log")
	}

	prevIndex := int(arg.PrevLogIndex)
	if prevIndex > len(logs)-1 || logs[prevIndex].Term != arg.PrevLogTerm {
		return res, nil
	}

	j := 0
	for i := prevIndex + 1; i <= len(logs)-1; i++ {
		if arg.Entries[j].Term != logs[i].Term {
			logs = logs[:i]
			break
		}
	}

	logs = append(logs, arg.Entries[j:]...)
	s.SetLog(logs)

	if int(arg.LeaderCommit) > s.commitIndex {
		s.commitIndex = min(int(arg.LeaderCommit), len(logs)-1)
	}
	res.Success = true
	return res, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Voting Logic
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
				s.TryStopTimeout()
				s.TryStartHeartbeat()
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
