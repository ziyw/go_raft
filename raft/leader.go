package raft

import (
	"context"
	_ "fmt"
	_ "github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	_ "os"
	_ "os/signal"
	"time"
)

func (s *Server) SendingHeartbeat(ctx context.Context) {
	log.Printf("Server %s: Start to send heartbeat\n", s.Addr)

	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		select {
		case <-s.StopHeartbeat:
			log.Printf("Server %s: Leader Action Stopped\n", s.Addr)
			return
		case <-ctx.Done():
			log.Printf("Server %s: context cancelled\n", s.Addr)
			return
		case <-ticker.C:
			for _, addr := range s.Cluster {
				if addr == s.Addr {
					continue
				}
				go s.SendHeartbeat(addr)
			}
		}
	}
}

func (s *Server) SendHeartbeat(addr string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// client := pb.NewRaftServiceClient(conn)

	select {
	case <-s.StopHeartbeat:
		return
	default:
		log.Printf("beat: [%s] to [%s]", s.Addr, addr)

		// actual reply feedback
		// 		arg := &pb.AppendArg{
		// 			Term: int64(s.CurrentTerm()),
		// 		}
		// 		reply, _ := client.AppendEntries(ctx, arg)
		// 		if reply.Term > int64(s.CurrentTerm()) {
		// 			log.Printf("Server %s: Change role from leader to follower\n", s.Addr)
		// 			s.stopHeartbeat <- struct{}{}
		// 		}
		return
	}
}
