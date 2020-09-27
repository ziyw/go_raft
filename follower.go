package main

import (
	"golang.org/x/net/context"
)

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	res := AppendRes{Term: s.currentTerm, Success: false}

	if arg.Term < s.currentTerm {
		// TODO: Trigger Vote
		return &res, nil
	}

	prevIndex := int(arg.PrevLogIndex)
	if prevIndex > len(s.log)-1 || s.log[prevIndex].Term != arg.PrevLogTerm {
		return &res, nil
	}

	j := 0
	for i := prevIndex + 1; i <= len(s.log)-1; i++ {
		if arg.Entries[j].Term != s.log[i].Term {
			s.log = s.log[:i]
			break
		}
		j++
	}

	s.log = append(s.log, arg.Entries[j:]...)

	if int(arg.LeaderCommit) > s.commitIndex {
		s.commitIndex = min(int(arg.LeaderCommit), len(s.log)-1)
	}

	if arg.Term > s.currentTerm {
		s.currentTerm = arg.Term
	}
	res.Success = true
	return &res, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
