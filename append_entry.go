package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func (s *Server) LeaderAction(followers []*Server, done chan int) {

	s.group = followers
	updateDone := make(chan int)
	for i := 0; i < len(followers); i++ {
		go s.SendAppend(i, updateDone)
	}

	go func() {
		count := 0
		for {
			select {
			case <-updateDone:
				count++
				if count > len(followers)/2 {
					done <- 1
				}
			}
		}
	}()
}

func (s *Server) SendAppend(target int, done chan int) {
	conn, err := grpc.Dial(s.group[target].Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewRaftServiceClient(conn)
	reply, err := client.AppendEntries(context.Background(), s.NewAppendArg(target))
	for !reply.Success || err != nil {
		s.nextIndex[target]--
		reply, err = client.AppendEntries(context.Background(), s.NewAppendArg(target))
	}
	done <- 1
}

func (s *Server) NewAppendArg(target int) *AppendArg {
	startIndex := int64(s.nextIndex[target])
	return &AppendArg{
		Term:         s.currentTerm,
		LeaderId:     int64(s.Id),
		PrevLogIndex: startIndex - 1,
		PrevLogTerm:  s.log[startIndex-1].Term,
		Entries:      s.log[startIndex:],
		LeaderCommit: int64(s.commitIndex),
	}

}

func (s *Server) SetLeader(followers []*Server) {
	s.nextIndex = make([]int, len(followers))
	s.matchIndex = make([]int, len(followers))

	for i := 0; i < len(followers); i++ {
		s.nextIndex[i] = len(s.log) + 1
		s.matchIndex[i] = 0
	}
}

// Follower behavior

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

	res.Success = true
	return &res, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
