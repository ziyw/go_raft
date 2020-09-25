package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func (s *Server) CandidateAction() {
	count := 0
	for _, f := range s.group {
		if f.Name == s.Name {
			continue
		}
		reply, err := s.SendVoteRequest(f, s.NewVote())
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Server %s Receive Reply %v\n", s.Name, reply)
		if reply.VoteGranted == true {
			count++
			log.Println("Add one")
		}
	}
}

// TODO: need to control when to stop sendVote
func (s *Server) SendVote(target int, done chan int) {
	conn, err := grpc.Dial(s.group[target].Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) SendVoteRequest(other *Server, req *VoteArg) (*VoteRes, error) {
	conn, err := grpc.Dial(other.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := NewRaftServiceClient(conn)
	return client.RequestVote(context.Background(), req)
}

func (s *Server) NewVote() *VoteArg {
	return &VoteArg{
		Term:         s.currentTerm,
		CandidateId:  int64(s.Id),
		LastLogIndex: int64(len(s.log) - 1),
		LastLogTerm:  s.currentTerm,
	}
}

func (s *Server) RequestVote(ctx context.Context, arg *VoteArg) (*VoteRes, error) {
	res := VoteRes{Term: s.currentTerm, VoteGranted: false}
	if arg.Term < s.currentTerm {
		return &res, nil
	}

	// TODO: compare if logs are update to date
	if s.votedFor == -1 || s.votedFor == arg.CandidateId {
		res.VoteGranted = true
		return &res, nil
	}
	return &res, nil
}
