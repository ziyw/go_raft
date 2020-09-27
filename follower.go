package main

import (
	"golang.org/x/net/context"
	"log"
)

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	term, err := s.CurrentTerm()
	if err != nil {
		log.Fatal(err)
	}
	res := AppendRes{Term: int64(term), Success: false}

	if arg.Term < int64(term) {
		// TODO: Trigger Vote
		return &res, nil
	}
	if arg.Term > int64(term) {
		s.SetCurrentTerm(int(arg.Term))
	}

	logs, err := s.Log()

	if err != nil {
		log.Fatal(err)
	}

	prevIndex := int(arg.PrevLogIndex)
	if prevIndex > len(logs)-1 || logs[prevIndex].Term != arg.PrevLogTerm {
		return &res, nil
	}

	j := 0
	for i := prevIndex + 1; i <= len(logs)-1; i++ {
		if arg.Entries[j].Term != logs[i].Term {
			logs = logs[:i]
			break
		}
		j++
	}

	logs = append(logs, arg.Entries[j:]...)

	s.SaveEntries(logs)

	if int(arg.LeaderCommit) > s.commitIndex {
		s.commitIndex = min(int(arg.LeaderCommit), len(logs)-1)
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
