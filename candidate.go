package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func (s *Server) InitCandidate() {
	s.State = Candidate
	cur, err := s.CurrentTerm()
	if err != nil {
		log.Fatal(err)
	}
	s.SetCurrentTerm(cur + 1)
}

func (s *Server) StartVote() {
	quit := make(chan struct{})
	done := make(chan int)
	for i := 0; i < len(s.group); i++ {
		go func(quit chan struct{}, index int) {
			select {
			case <-quit:
				return
			default:
				s.SendVote(index, done)
			}
		}(quit, i)
	}

	count := 0
	for i := range done {
		count += i
		if count >= len(s.group)/2 {
			// Change State
		}
		quit <- struct{}{}
		return
	}
}

func (s *Server) SendVote(target int, done chan int) {
	conn, err := grpc.Dial(s.group[target].Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := NewRaftServiceClient(conn)
	r, err := client.RequestVote(context.Background(), s.NewVote())
	if err != nil {
		log.Fatal(err)
	}
	if r.VoteGranted {
		done <- 1
	}
	done <- 0
}

func (s *Server) NewVote() *VoteArg {
	cur, err := s.CurrentTerm()
	if err != nil {
		log.Fatal(err)
	}
	l, err := s.Log()
	if err != nil {
		log.Fatal(err)
	}
	return &VoteArg{
		Term:         int64(cur),
		CandidateId:  int64(s.Id),
		LastLogIndex: int64(len(l) - 1),
		LastLogTerm:  int64(cur),
	}
}

func (s *Server) RequestVote(ctx context.Context, arg *VoteArg) (*VoteRes, error) {
	term, err := s.CurrentTerm()
	if err != nil {
		log.Fatal(err)
	}
	res := VoteRes{Term: int64(term), VoteGranted: false}
	if arg.Term < int64(term) {
		return &res, nil
	}

	votedFor, err := s.VotedFor()
	if err != nil {
		log.Fatal(err)
	}
	if votedFor == -1 || votedFor == int(arg.CandidateId) {
		s.SetVotedFor(int(arg.CandidateId))
		res.VoteGranted = true
		return &res, nil
	}

	logs, err := s.Log()
	if err != nil {
		log.Fatal(err)
	}
	lastIndex := len(logs) - 1
	lastTerm := logs[lastIndex].Term

	if lastIndex == int(arg.LastLogIndex) && lastTerm == arg.LastLogTerm {
		s.SetVotedFor(int(arg.CandidateId))
		res.VoteGranted = true
		return &res, nil
	}
	return &res, nil
}
