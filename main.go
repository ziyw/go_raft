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

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	res := AppendRes{Term: s.currentTerm, Success: true}
	// rule 1:
	if arg.Term < s.currentTerm {
		res.Success = false
		return &res, nil
	}

	// rule 2:
	if s.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
		res.Success = false
		return &res, nil
	}

	// rule 3:Q: how to tell the difference of indexes
	for index, entry := range arg.Entries {
		if s.log[index].Term != entry.Term {
			s.log = s.log[:index]
		}
	}

	// rule 4: // TODO: make this two transpotable)
	// s.log = append(s.log, arg.Entries)

	// rule 5:
	// if arg.LeaderCommit > s.commitedIndex {
	//		if arg.LeaderCommit > len(s.log) {
	//			s.commitIndex = arg.LeaderCommited
	//		}
	//	}
	return &res, nil
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

	s1.log = make([]Entry, 10)
	s2.log = make([]Entry, 10)

	go s1.Start()
	go s2.Start()

	s1.SendAppendRequest(s2)

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
