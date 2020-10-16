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

func (s *Server) StartHeartbeat(ctx context.Context, cancel context.CancelFunc) {
	log.Println("Start to sending heartbeat")

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				for _, f := range s.followers {
					go s.SendHeartbeat(ctx, cancel, f.Addr)
				}
			}
		}
	}()

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

func (s *Server) SendHeartbeat(ctx context.Context, cancel context.CancelFunc, addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewRaftServiceClient(conn)

	select {
	case <-ctx.Done():
		log.Println("Server %s: context cancelled\n", s.Addr)
		return
	default:
		log.Printf("Server %s: send heartbeat to %s\n", s.Addr, addr)
		arg := &pb.AppendArg{
			Term: int64(s.CurrentTerm()),
		}
		reply, _ := client.AppendEntries(ctx, arg)
		if reply.Term > int64(s.CurrentTerm()) {
			s.StopAll <- struct{}{}
		}
		return
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

func (s *Server) HandleQuery(ctx context.Context, cancel context.CancelFunc, arg *pb.QueryArg) (*pb.QueryRes, error) {
	log.Printf("Server %s receive request %v\n", s.Addr, arg)
	s.AppendLog(&pb.Entry{Term: int64(s.CurrentTerm()), Command: arg.Command})

	logs := s.Log()
	for i := 0; i < len(s.followers); i++ {
		s.nextIndex[i] = len(logs)
	}

	// Start send log job
	done := make(chan *pb.AppendRes)
	for i, f := range s.followers {
		go s.sendAppendRequest(ctx, cancel, i, f, done)
	}

	// Collect result
	count := 0
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("Context cancelled")
		case r := <-done:
			if r.Term > int64(s.CurrentTerm()) {
				cancel()
			}
			if r.Success {
				count++
			}
			if count >= len(s.followers) {
				return &pb.QueryRes{Success: true, Reply: "Done"}, nil
			}
		}
	}
}

func (s *Server) sendAppendRequest(ctx context.Context, cancel context.CancelFunc, idx int, follower *Server, done chan *pb.AppendRes) {
	conn, err := grpc.Dial(s.followers[idx].Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewRaftServiceClient(conn)
	r, err := client.AppendEntries(ctx, s.NewAppendArg(idx))

	for {
		if r.Success {
			done <- r
			return
		}
		select {
		case <-s.StopAll:
			cancel()
			return
		case <-ctx.Done():
			log.Printf("Server %s: cancelled", s.Addr)
			return
		default:
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
