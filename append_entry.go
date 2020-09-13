package main

import "golang.org/x/net/context"

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	res := AppendRes{Term: s.currentTerm, Success: true}
	// rule 1:
	if arg.Term < s.currentTerm {
		res.Success = false
		return &res, nil
	}

	// rule 2:
	if s.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
		res.Success = false
		return &res, nil
	}

	// rule 3:Q: how to tell the difference of indexes
	for index, entry := range arg.Entries {
		if s.log[index].Term != entry.Term {
			s.log = s.log[:index]
		}
	}

	// rule 4: // TODO: make this two transpotable)
	// s.log = append(s.log, arg.Entries)

	// rule 5:
	// if arg.LeaderCommit > s.commitedIndex {
	//		if arg.LeaderCommit > len(s.log) {
	//			s.commitIndex = arg.LeaderCommited
	//		}
	//	}
	return &res, nil
}
