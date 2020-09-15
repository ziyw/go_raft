package main

import "golang.org/x/net/context"

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	res := AppendRes{Term: s.currentTerm, Success: true}

	if arg.Term < s.currentTerm {
		res.Success = false
		return &res, nil
	}

	if s.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
		res.Success = false
		return &res, nil
	}

	i := 0
	for i < len(arg.Entries) {
		index := int(i + int(arg.PrevLogIndex) + 1)
		if s.log[index].Term != arg.Entries[i].Term {
			s.log = s.log[index-1:]
			break
		}
	}

	for _, v := range arg.Entries {
		s.log = append(s.log, *v)
	}

	// rule 5:
	if arg.LeaderCommit > s.commitIndex {
		if int(arg.LeaderCommit) > int(len(s.log)) {
			s.commitIndex = uint32(arg.LeaderCommit)
		} else {
			s.commitIndex = uint32(len(s.log))
		}
	}
	return &res, nil
}
