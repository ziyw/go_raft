package main

import "golang.org/x/net/context"

func (s *Server) RequestVote(ctx context.Context, arg *VoteArg) (*VoteRes, error) {
	return &VoteRes{}, nil
}
