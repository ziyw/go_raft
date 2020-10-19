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

func (s *Server) TryStartHeartbeat() {
	if s.HeartbeatRunning {
		return
	}
	s.StartHeartbeat <- struct{}{}
}

func (s *Server) TryStopHeartbeat() {
	if !s.HeartbeatRunning {
		return
	}
	s.StopHeartbeat <- struct{}{}
}

func (s *Server) SendingHeartbeat(ctx context.Context) {
	log.Printf("Server %s: Start to send heartbeat\n", s.Addr)
	s.HeartbeatRunning = true

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		select {
		case <-s.StopHeartbeat:
			s.HeartbeatRunning = false
			log.Printf("Server %s: Leader Action Stopped\n", s.Addr)
			return
		case <-ctx.Done():
			s.HeartbeatRunning = false
			log.Printf("Server %s: context cancelled\n", s.Addr)
			return
		case <-ticker.C:
			for _, addr := range s.Cluster {
				if addr == s.Addr {
					continue
				}
				go s.SendHeartbeat(ctx, addr)
			}
		}
	}
}

func (s *Server) SendHeartbeat(ctx context.Context, addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewRaftServiceClient(conn)

	select {
	case <-s.StopHeartbeat:
		return
	default:
		log.Printf("beat: [%s] to [%s]", s.Addr, addr)
		arg := &pb.AppendArg{
			Term: int64(s.CurrentTerm()),
		}
		reply, _ := client.AppendEntries(ctx, arg)
		if reply.Term > int64(s.CurrentTerm()) {
			log.Printf("Server %s: Change role from leader to follower\n", s.Addr)
			s.TryStopHeartbeat()
		}
		return
	}
}

// Handle query from client logic.
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
		case <-s.StopHeartbeat:
			log.Println("Leader is stopped")
			return nil, fmt.Errorf("Leader error")
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
			s.TryStopHeartbeat()
			return
		}

		select {
		case <-s.StopHeartbeat:
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
