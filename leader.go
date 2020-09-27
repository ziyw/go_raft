package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	_ "net"
	_ "os"
	_ "strconv"
	_ "strings"
)

// Query is receive normal query from normal client.
func (s *Server) Query(ctx context.Context, arg *QueryArg) (*QueryRes, error) {
	if s.State == Leader {
		r := fmt.Sprintf("%s says Hi", s.Name)
		t, err := s.CurrentTerm()
		if err != nil {
			return nil, nil
		}
		s.SaveEntry(&Entry{Term: int64(t), Command: arg.Command})
		return &QueryRes{Success: true, Reply: r}, nil
	}

	l, err := s.VotedFor()
	if err != nil || l > len(s.group) {
		l = -1
		for i, v := range s.group {
			if v.State == Leader {
				l = i
			}
		}
	}

	if l == -1 {
		// TODO: trigger VoteAtion here
		return nil, fmt.Errorf("No valid leader in the group")
	}

	curLeader := s.group[l]
	conn, err := grpc.Dial(curLeader.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := NewRaftServiceClient(conn)
	return c.Query(context.Background(), arg)
}

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
