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
		return s.HandleQuery(arg)
	}

	conn, err := grpc.Dial(s.leaderAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := NewRaftServiceClient(conn)
	return c.Query(context.Background(), arg)
}

func (s *Server) HandleQuery(arg *QueryArg) (*QueryRes, error) {
	quit := make(chan struct{})
	done := make(chan int)

	for i := 0; i < len(s.group); i++ {
		go func(quit chan struct{}, index int) {
			select {
			case <-quit:
				return
			default:
				s.SendAppend(index, done)
			}
		}(quit, i)
	}

	count := 0
	for i := range done {
		count += i
		if count >= len(s.group)/2 {
			return &QueryRes{Success: true, Reply: "Hello"}, nil
		}
	}
	return nil, fmt.Errorf("Lack of result")
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
	term, err := s.CurrentTerm()
	if err != nil {
		log.Fatal(err)
	}

	logs, err := s.Log()
	if err != nil {
		log.Fatal(err)
	}

	return &AppendArg{
		Term:         int64(term),
		LeaderId:     int64(s.Id),
		PrevLogIndex: startIndex - 1,
		PrevLogTerm:  logs[startIndex-1].Term,
		Entries:      logs[startIndex:],
		LeaderCommit: int64(s.commitIndex),
	}

}

func (s *Server) InitLeader(followers []*Server) {
	s.State = Leader
	s.nextIndex = make([]int, len(followers))
	s.matchIndex = make([]int, len(followers))

	logs, err := s.Log()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(followers); i++ {
		s.nextIndex[i] = len(logs) + 1
		s.matchIndex[i] = 0
	}
}

func (s *Server) InitFollower(newTerm int) {
	s.State = Follower
	s.SetCurrentTerm(newTerm)
}
