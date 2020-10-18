package raft

import (
	"context"
	"fmt"
	"github.com/ziyw/go_raft/pb"
	_ "google.golang.org/grpc"
	_ "log"
	_ "time"
)

func (s *Server) HandleAppendEntries(ctx context.Context, arg *pb.AppendArg) (*pb.AppendRes, error) {
	s.stopElectionTimeout <- struct{}{}
	term := int64(s.CurrentTerm())
	if term < 0 {
		err := fmt.Errorf("Invalid term number")
		return nil, err
	}

	res := &pb.AppendRes{Term: term, Success: false}

	if arg.Term < term {
		return res, fmt.Errorf("Leader out-of-date")
	}

	if arg.Term > term {
		if s.Role == "l" {
			s.stopHeartbeat <- struct{}{}
			return nil, fmt.Errorf("Leader out of date")
		}
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
