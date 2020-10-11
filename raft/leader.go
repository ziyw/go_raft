package raft

import (
	"context"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	"time"
)

// Proto buffer Impl
func (s *Server) Query(ctx context.Context, arg *pb.QueryArg) (*pb.QueryRes, error) {
	return nil, nil
}

func (s *Server) AppendEntries(ctx context.Context, arg *pb.AppendArg) (*pb.AppendRes, error) {
	return nil, nil
}

func (s *Server) RequestVote(ctx context.Context, arg *pb.VoteArg) (*pb.VoteRes, error) {
	return nil, nil
}

func (s *Server) Heatbeat(ctx context.Context, arg int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, f := range s.followers {
				go s.SendAppendEmpty(ctx, f)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Server) SendAppendEmpty(ctx context.Context, target *Server) {
	conn, err := grpc.Dial(target.Addr, grpc.WithInsecure())
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
		ctx.Done()
	}
}
