package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"os"
)

type Server struct {
	Name string
	Addr string

	// leader state
	nextIndex  []uint32
	matchIndex []uint32

	// common server state
	commitedIndex uint32
	lastApplied   uint32
	// persist states
	currentTerm uint32
	votedFor    uint32
	log         []LogEntry
}

type LogEntry struct {
	command string
	term    uint32
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

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {

	// TODO: insert append logic
	return &AppendRes{Term: 1, Success: true}, nil
}

func (s *Server) RequestVote(ctx context.Context, arg *VoteArg) (*VoteRes, error) {
	return &VoteRes{}, nil
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
	log.Println("Reply %s", r.Term)
}

func main() {

	s1 := Server{Name: "NodeOne", Addr: "localhost:60001"}
	s2 := Server{Name: "NodeTwo", Addr: "localhost:60002"}

	go s1.Start()
	go s2.Start()

}

func fileIo() {
	f, err := os.Create("data.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data := []byte("HelloWorld\n")
	err = ioutil.WriteFile("data", data, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Println("Finish")

}
