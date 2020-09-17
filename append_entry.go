package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// func (s *Server) SendAppendRequest(other Server, req AppendArg) (*AppendRes, error) {
// 	conn, err := grpc.Dial(other.Addr, grpc.WithInsecure())
// 	Check(err)
// 	defer conn.Close()
//
// 	client := NewRaftServiceClient(conn)
// 	return client.AppendEntries(context.Background(), &req)
// }
//
func (s *Server) SendAppendRequest(servers *[]Server) {
	followers := *servers
	for i, f := range followers {
		if len(s.log) < int(s.nextIndex[i]) {
			continue
		}

		conn, err := grpc.Dial(f.Addr, grpc.WithInsecure())
		Check(err)
		defer conn.Close()
		client := NewRaftServiceClient(conn)

		request := AppendArg{
			Term:         s.currentTerm,
			PrevLogIndex: int64(s.nextIndex[i]),
			PrevLogTerm:  s.log[s.nextIndex[i]].Term,
			Entries:      s.log[s.nextIndex[i]:],
			LeaderCommit: int64(s.commitIndex),
		}

		reply, _ := client.AppendEntries(context.Background(), &request)
		for {
			if reply.Success {
				s.nextIndex[i] = s.commitIndex + 1
				s.matchIndex[i] = s.commitIndex
				break
			}

			s.nextIndex[i]--

			request := AppendArg{
				Term:         s.currentTerm,
				PrevLogIndex: int64(s.nextIndex[i]),
				PrevLogTerm:  s.log[s.nextIndex[i]].Term,
				Entries:      s.log[s.nextIndex[i]:],
				LeaderCommit: int64(s.commitIndex),
			}
			reply, _ = client.AppendEntries(context.Background(), &request)
		}

	}
}

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	res := AppendRes{Term: s.currentTerm, Success: true}

	if arg.Term < s.currentTerm {
		res.Success = false
		return &res, nil
	}

	// TODO: this is wrong check later
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

	s.log = append(s.log, arg.Entries...)

	// TODO: this is not correct
	if arg.LeaderCommit > int64(s.commitIndex) {
		if int(arg.LeaderCommit) > int(len(s.log)) {
			s.commitIndex = int(arg.LeaderCommit)
		} else {
			s.commitIndex = int(len(s.log))
		}
	}
	return &res, nil
}
