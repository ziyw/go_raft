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

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.StopLead:
			log.Printf("Server %s: Leader Action Stopped\n", s.Addr)
			cancel()
			return
		case <-ctx.Done():
			log.Printf("Server %s: context cancelled\n", s.Addr)
			return
		case <-ticker.C:
			for _, addr := range s.Cluster {
				if addr == s.Addr {
					continue
				}
				go s.SendHeartbeat(ctx, cancel, addr)
			}
		}
	}
}

func (s *Server) StopHeartbeat() {
	s.StopLead <- struct{}{}
}

func (s *Server) SendHeartbeat(ctx context.Context, cancel context.CancelFunc, addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewRaftServiceClient(conn)

	select {
	case <-s.StopLead:
		cancel()
		return
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
			s.StopLead <- struct{}{}
			s.StartFollow <- struct{}{}
		}
		return
	}
}

// Reply to vote request, grant vote if qualified.
func (s *Server) HandleRequestVote(ctx context.Context, arg *pb.VoteArg) (*pb.VoteRes, error) {
	term := int64(s.CurrentTerm())
	res := &pb.VoteRes{Term: term, VoteGranted: false}

	if arg.CandidateId == s.Addr {
		res.VoteGranted = true
		return res, nil
	}

	if arg.Term < term {
		s.StopLead <- struct{}{}
		s.StartFollow <- struct{}{}
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
		s.StopLead <- struct{}{}
		s.StartFollow <- struct{}{}
		return res, nil
	}
	return res, nil
}

// TODO: making sure only one leader exist
func (s *Server) HandleQuery(ctx context.Context, cancel context.CancelFunc, arg *pb.QueryArg) (*pb.QueryRes, error) {
	log.Printf("Server %s receive request %v\n", s.Addr, arg)
	s.AppendLog(&pb.Entry{Term: int64(s.CurrentTerm()), Command: arg.Command})

	logs := s.Log()
	for i := 0; i < len(s.Cluster); i++ {
		s.nextIndex[i] = len(logs)
	}

	// Start send log job
	done := make(chan *pb.AppendRes)
	for i, addr := range s.Cluster {
		if addr == s.Addr {
			continue
		}
		go s.sendAppendRequest(ctx, cancel, i, done)
	}

	// Collect result
	count := 0
	for {
		select {
		case <-s.StopLead:
			return nil, fmt.Errorf("Leader is stopped")
		case <-ctx.Done():
			return nil, fmt.Errorf("Context cancelled")
		case r := <-done:
			if r.Term > int64(s.CurrentTerm()) {
				cancel()
			}
			if r.Success {
				count++
			}
			if count >= len(s.Cluster)/2+1 {
				return &pb.QueryRes{Success: true, Reply: "Done"}, nil
			}
		}
	}
}

func (s *Server) sendAppendRequest(ctx context.Context, cancel context.CancelFunc, idx int, done chan *pb.AppendRes) {
	conn, err := grpc.Dial(s.Cluster[idx], grpc.WithInsecure())
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
		if r.Term > int64(s.CurrentTerm()) {
			s.StopLead <- struct{}{}
			return
		}

		select {
		case <-s.StopLead:
			cancel()
			return
		case <-ctx.Done():
			log.Printf("Server %s: cancelled", s.Addr)
			return
		default:
			s.nextIndex[idx]--
			r, err = client.AppendEntries(ctx, s.NewAppendArg(idx))
		}

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
