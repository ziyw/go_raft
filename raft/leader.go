package raft

import (
	"context"
	"fmt"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	_ "os"
	_ "os/signal"
	"time"
)

func (s *Server) InitLeader(ctx context.Context) {
	s.nextIndex = make([]int, len(s.followers))
	s.matchIndex = make([]int, len(s.followers))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	sendEmpty := func(ctx context.Context, f *Server) {
		log.Printf("Leader %s send heartbeat\n", s.Addr)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		conn, err := grpc.Dial(f.Addr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		term := s.CurrentTerm()
		arg := &pb.AppendArg{
			Term:    int64(term),
			Entries: []*pb.Entry{},
		}

		client := pb.NewRaftServiceClient(conn)
		reply, err := client.AppendEntries(ctx, arg)
		if reply.Term > int64(term) {
			cancel()
			ctx.Done()
		}
	}

	for {
		select {
		case <-ticker.C:
			for _, f := range s.followers {
				go sendEmpty(ctx, f)
			}
		case <-ctx.Done():
			log.Fatal(ctx.Err())
		}
	}

}

// Proto buffer Impl
func (s *Server) HandleQuery(ctx context.Context, arg *pb.QueryArg) (*pb.QueryRes, error) {
	log.Printf("Receive query request %v\n", arg)

	logs := s.Log()
	for i := 0; i < len(s.followers); i++ {
		s.nextIndex[i] = len(logs)
	}
	// TODO: add signal later
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan int)
	for i := 0; i < len(s.followers); i++ {
		go s.SendAppendEntries(ctx, done, i)
	}

	cnt := 0
	for i := range done {
		cnt += i
		if cnt > len(s.followers)/2 {
			s.AppendLog(&pb.Entry{Term: int64(s.CurrentTerm()), Command: arg.Command})
			return &pb.QueryRes{Success: true, Reply: "Done"}, nil
		}
	}
	return nil, fmt.Errorf("Handle Query Error")
}

func (s *Server) SendAppendEntries(ctx context.Context, done chan int, idx int) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := grpc.Dial(s.followers[idx].Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	term := s.CurrentTerm()
	client := pb.NewRaftServiceClient(conn)
	r, err := client.AppendEntries(ctx, s.NewAppendArg(idx))

	for {
		if r.Term > int64(term) {
			cancel()
			// TODO: also cancel parent
		}

		if r.Success {
			done <- 1
			return
		}
		s.nextIndex[idx]--
		r, err = client.AppendEntries(ctx, s.NewAppendArg(idx))
	}
}

func (s *Server) NewAppendArg(target int) *pb.AppendArg {
	startIndex := s.nextIndex[target]
	logs := s.Log()
	return &pb.AppendArg{
		Term:         int64(s.CurrentTerm()),
		LeaderId:     s.Id,
		PrevLogIndex: int64(startIndex - 1),
		PrevLogTerm:  logs[startIndex-1].Term,
		Entries:      logs[startIndex:],
		LeaderCommit: int64(s.commitIndex),
	}
}

// Reply to vote request, grant vote if qualified.
func (s *Server) HandleRequestVote(ctx context.Context, arg *pb.VoteArg) (*pb.VoteRes, error) {
	term := int64(s.CurrentTerm())
	res := &pb.VoteRes{Term: term, VoteGranted: false}
	if arg.Term < term {
		return res, nil
	}

	votedFor := s.VotedFor()
	if votedFor == "" || votedFor == arg.CandidateId {
		s.SetVotedFor(arg.CandidateId)
		res.VoteGranted = true
		return res, nil
	}

	log := s.Log()
	lastIndex := len(log) - 1
	lastTerm := log[lastIndex].Term
	if lastIndex == int(arg.LastLogIndex) && lastTerm == arg.LastLogTerm {
		s.SetVotedFor(arg.CandidateId)
		res.VoteGranted = true
		return res, nil
	}
	return res, nil
}
