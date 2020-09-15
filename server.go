package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Server struct {
	Name string
	Addr string

	// leader state
	nextIndex  []uint32
	matchIndex []uint32

	// common server state
	commitIndex uint32
	lastApplied uint32

	// persist states, TODO: need to write those to file
	currentTerm uint32
	votedFor    uint32
	log         []Entry
}

func (s *Server) Start() {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	_s := grpc.NewServer()
	RegisterRaftServiceServer(_s, s)
	if err := _s.Serve(lis); err != nil {
		log.Fatalf("Failed to server %v", err)
	}
}

func (s *Server) SendAppendRequest(other Server) {
	conn, err := grpc.Dial(other.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Connection error %v", err)
	}
	defer conn.Close()

	client := NewRaftServiceClient(conn)
	req := AppendArg{}
	r, err := client.AppendEntries(context.Background(), &req)
	if err != nil {
		log.Fatalf("Service Failed %v", err)
	}

	log.Println("Receive respond: ", r.Term, r.Success)
}

// TODO
func (s *Server) SendRequestVote(other Server) {}

// TODO
func (s Server) IsLeader() bool { return false }
