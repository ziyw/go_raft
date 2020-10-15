package raft

import (
	"context"
	_ "fmt"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	_ "os"
	_ "os/signal"
	"time"
)

func (s *Server) StartHeartbeat(ctx context.Context, cancel context.CancelFunc) {
	log.Println("Start to sending heartbeat")
	for _, f := range s.followers {
		go s.KeepSending(ctx, cancel, f.Addr)
	}
	select {
	case <-s.StopAll:
		log.Printf("Server %s: Stop all\n", s.Addr)
		cancel()
		return
	case <-ctx.Done():
		log.Printf("Server %s: context cancelled\n", s.Addr)
		return
	}
}

func (s *Server) StopHeartbeat() {
	s.StopAll <- struct{}{}
}

func (s *Server) KeepSending(ctx context.Context, cancel context.CancelFunc, addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewRaftServiceClient(conn)
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-s.StopAll:
			cancel()
			return
		case <-ctx.Done():
			log.Printf("Server %s: context cancelled\n", s.Addr)
			return
		case <-ticker.C:
			log.Printf("Server %s: send heartbeat to %s\n", s.Addr, addr)
			arg := &pb.AppendArg{
				Term:    int64(s.CurrentTerm()),
				Entries: []*pb.Entry{},
			}
			reply, _ := client.AppendEntries(ctx, arg)
			if reply.Term > int64(s.CurrentTerm()) {
				s.StopAll <- struct{}{}
			}
		}
	}
}

// TODO: This is not working yet.
// Handle client sending query to leader.
func (s *Server) HandleQuery(ctx context.Context, arg *pb.QueryArg) (*pb.QueryRes, error) {
	log.Printf("Receive query request %v\n", arg)

	logs := s.Log()
	for i := 0; i < len(s.followers); i++ {
		s.nextIndex[i] = len(logs)
	}

	done := make(chan struct{})
	for i := 0; i < len(s.followers); i++ {
		go s.SendAppendEntries(ctx, done, i)
	}

	applied := 0
	for {
		select {
		case <-ctx.Done():
			log.Println("Handle query cancelled")
			return &pb.QueryRes{Success: false, Reply: "Error"}, nil
		case <-done:
			if applied > len(s.followers)/2 {
				s.AppendLog(&pb.Entry{Term: int64(s.CurrentTerm()), Command: arg.Command})
				return &pb.QueryRes{Success: true, Reply: "Done"}, nil
			}
		}
	}

	return &pb.QueryRes{Success: false, Reply: "Not done"}, nil
}

func (s *Server) SendAppendEntries(ctx context.Context, done chan struct{}, idx int) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := grpc.Dial(s.followers[idx].Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewRaftServiceClient(conn)
	r, err := client.AppendEntries(ctx, s.NewAppendArg(idx))

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.StopAll:
			cancel()
			return
		default:
		}
		if r.Success && r.Term <= int64(s.CurrentTerm()) {
			done <- struct{}{}
			return
		}

		// When getting term bigger than currentTerm, leader should stop everything
		// and becomes follower next
		if r.Term > int64(s.CurrentTerm()) {
			s.StopAll <- struct{}{}
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
