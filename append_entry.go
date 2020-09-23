package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func (s *Server) LeaderAction(servers *[]Server, done chan int) {
	for i, f := range s.group {
		if f.Name == s.Name {
			continue
		}

		r, err := s.SendAppendRequest(f, s.NewAppendArg(i))
		if err != nil {
			log.Fatal(err)
		}

		for !r.Success && s.nextIndex[i] >= 0 {
			s.nextIndex[i]--
			r, err = s.SendAppendRequest(f, s.NewAppendArg(i))
			if err != nil {
				log.Fatal(err)
			}
		}

		if r.Success {
			s.nextIndex[i] = s.commitIndex + 1
			s.matchIndex[i] = s.commitIndex
			continue
		}
	}
}

func (s *Server) SendAppendRequest(to *Server, req *AppendArg) (*AppendRes, error) {
	conn, err := grpc.Dial(to.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := NewRaftServiceClient(conn)
	return c.AppendEntries(context.Background(), req)
}

func (s *Server) InitLeaderState() {
	s.nextIndex = make([]int, len(s.group))
	s.matchIndex = make([]int, len(s.group))

	for i := 0; i < len(s.group); i++ {
		s.nextIndex[i] = len(s.log) - 1
		s.matchIndex[i] = 0
	}
}

func (s *Server) NewAppendArg(index int) *AppendArg {
	return &AppendArg{
		Term:         s.currentTerm,
		PrevLogIndex: int64(s.nextIndex[index]),
		PrevLogTerm:  s.log[s.nextIndex[index]].Term,
		Entries:      s.log[s.nextIndex[index]:],
		LeaderCommit: int64(s.commitIndex),
	}
}

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	log.Printf("Server %s Receive AppendEntries Request %v\n", s.Name, *arg)

	res := AppendRes{Term: s.currentTerm, Success: true}

	if arg.Term < s.currentTerm {
		res.Success = false
		return &res, nil
	}

	if len(s.log) == 0 {
		res.Success = true
		s.log = arg.Entries
		return &res, nil
	}

	if len(s.log) <= int(arg.PrevLogIndex) {
		res.Success = false
		return &res, nil
	}

	if s.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
		res.Success = false
		return &res, nil
	}

	// check conflit
	for i := 0; i < len(arg.Entries); i++ {
		j := int(i + int(arg.PrevLogIndex))
		if j >= len(s.log) {
			break
		}
		if s.log[j].Term != arg.Entries[i].Term {
			s.log = s.log[:j-1]
			break
		}
	}

	// Append entries that are not exsiting
	s.log = append(s.log, arg.Entries...)

	if int(arg.LeaderCommit) > s.commitIndex {
		s.commitIndex = min(int(arg.LeaderCommit), len(s.log)-1)
	}

	return &res, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
