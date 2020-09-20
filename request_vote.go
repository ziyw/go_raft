package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func (s *Server) SendVoteRequest(other Server, req VoteArg) (*VoteRes, error) {
	conn, err := grpc.Dial(other.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	client := NewRaftServiceClient(conn)
	return client.RequestVote(context.Background(), &req)
}

func (s *Server) Vote(servers *[]Server, done chan int) {
	log.Printf("Start Voting \n")
	p := *servers
	for i := 0; i < len(p); i++ {
		conn, err := grpc.Dial(p[i].Addr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		client := NewRaftServiceClient(conn)

		request := VoteArg{
			Term:         s.currentTerm,
			CandidateId:  int64(s.Id),
			LastLogIndex: int64(len(s.log) - 1),
			LastLogTerm:  s.currentTerm,
		}
		reply, _ := client.RequestVote(context.Background(), &request)
		log.Printf("Server %s Receive Reply %v\n", s.Name, reply)
	}
	done <- 1
}

func (s *Server) RequestVote(ctx context.Context, arg *VoteArg) (*VoteRes, error) {
	res := VoteRes{Term: s.currentTerm, VoteGranted: false}
	if arg.Term < s.currentTerm {
		return &res, nil
	}

	// TODO: compare if logs are update to date
	if s.votedFor == -1 || s.votedFor == arg.CandidateId {
		res.VoteGranted = true
		return &res, nil
	}
	return &res, nil
}
