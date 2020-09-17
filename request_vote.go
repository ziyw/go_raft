package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func (s *Server) SendVoteRequest(other Server, req VoteArg) (*VoteRes, error) {
	conn, err := grpc.Dial(other.Addr, grpc.WithInsecure())
	Check(err)
	client := NewRaftServiceClient(conn)
	return client.RequestVote(context.Background(), &req)
}

func (s *Server) RequestVote(ctx context.Context, arg *VoteArg) (*VoteRes, error) {
	res := VoteRes{Term: s.currentTerm, VoteGranted: false}
	if arg.Term < s.currentTerm {
		return &res, nil
	}

	// TODO: this implementation is not full
	if s.votedFor == -1 || s.votedFor == arg.CandidateId {
		res.VoteGranted = true
		return &res, nil
	}
	return &res, nil
}
