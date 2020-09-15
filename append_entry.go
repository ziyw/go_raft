package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func (s *Server) SendAppendRequest(other Server, req AppendArg) (*AppendRes, error) {
	conn, err := grpc.Dial(other.Addr, grpc.WithInsecure())
	Check(err)
	defer conn.Close()

	client := NewRaftServiceClient(conn)
	return client.AppendEntries(context.Background(), &req)
}

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	res := AppendRes{Term: s.currentTerm, Success: true}

	if arg.Term < s.currentTerm {
		res.Success = false
		return &res, nil
	}

	if len(s.log) == 0 {
		return &res, nil
	}

	if s.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
		res.Success = false
		return &res, nil
	}

	i := 0
	for i < len(arg.Entries) {
		index := int(i + int(arg.PrevLogIndex) + 1)
		if index > len(s.log) {
			break
		}
		if s.log[index].Term != arg.Entries[i].Term {
			s.log = s.log[index-1:]
			break
		}
	}

	for _, v := range arg.Entries {
		s.log = append(s.log, *v)
	}

	// TODO: this is not correct
	if arg.LeaderCommit > s.commitIndex {
		if int(arg.LeaderCommit) > int(len(s.log)) {
			s.commitIndex = uint32(arg.LeaderCommit)
		} else {
			s.commitIndex = uint32(len(s.log))
		}
	}
	return &res, nil
}
