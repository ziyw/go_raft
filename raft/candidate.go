package raft

import (
	"context"
	_ "fmt"
	"github.com/ziyw/go_raft/pb"
	_ "google.golang.org/grpc"
	"log"
	_ "os"
	_ "os/signal"
	_ "time"
)

func (s *Server) HandleRequestVote(ctx context.Context, arg *pb.VoteArg) (*pb.VoteRes, error) {
	term := int64(s.CurrentTerm())
	res := &pb.VoteRes{Term: term, VoteGranted: false}

	if arg.Term < term {
		return res, nil
	}

	if arg.Term > term && s.Role == "l" {
		log.Printf("Server %s: switch role from leader to follower", s.Addr)
		s.SetCurrentTerm(int(arg.Term))
		s.TryStopHeartbeat()
		s.TryStartTimeout()
		res.VoteGranted = true
		return res, nil
	}

	if arg.CandidateId == s.Addr {
		res.VoteGranted = true
		return res, nil
	}

	votedFor := s.VotedFor()
	validVote := votedFor == "" || votedFor == arg.CandidateId

	log := s.Log()
	lastIndex := len(log) - 1
	lastTerm := log[lastIndex].Term
	validLog := lastIndex == int(arg.LastLogIndex) && lastTerm == arg.LastLogTerm

	if validVote && validLog {
		res.VoteGranted = true
		return res, nil
	}
	return res, nil
}
