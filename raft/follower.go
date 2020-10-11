package raft

import (
	"context"
	_ "github.com/ziyw/go_raft/pb"
	_ "google.golang.org/grpc"
	"log"
	"time"
)

func (s *Server) ListenHeartbeat(ctx context.Context, arg int) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Second):
		log.Println("Should start voting now")
		// start Vote
	}
}
