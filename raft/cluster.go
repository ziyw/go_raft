package raft

import (
	"context"
	_ "fmt"
	"github.com/ziyw/go_raft/file"
	"github.com/ziyw/go_raft/pb"
	"google.golang.org/grpc"
	"log"
	"strings"
	_ "time"
)

func NewCluster(cfg string) {
	lines, err := file.ReadLines(cfg)
	if err != nil {
		panic(err)
	}

	servers := []*Server{}
	var leader *Server
	followers := []*Server{}

	for _, l := range lines {
		cur := strings.Trim(l, "\n")
		if len(cur) == 0 {
			continue
		}
		arg := strings.Split(cur, ",")
		for _, a := range arg {
			a = strings.Trim(a, " ")
		}
		s := NewServer(arg[0], arg[1], arg[2], arg[3])
		servers = append(servers, s)
		if s.Role == "f" {
			followers = append(followers, s)
		} else {
			leader = s
		}
	}

	for _, s := range servers {
		s.leader = leader
		s.followers = followers
		ctx, cancel := context.WithCancel(context.Background())
		go s.Start(ctx, cancel)
	}

}

func SendRequest(addr string, body string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewRaftServiceClient(conn)
	req := &pb.QueryArg{Command: body}
	r, err := c.Query(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Reply ", r)
}
