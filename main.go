package main

import (
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	_ "sync"
	"time"
)

func main() {
	allDone := make(chan struct{})

	s1 := NewServer("s1", "localhost:30001", 1)
	s2 := NewServer("s2", "localhost:30002", 2)
	s3 := NewServer("s3", "localhost:30003", 3)
	s4 := NewServer("s4", "localhost:30004", 4)
	s5 := NewServer("s5", "localhost:30005", 5)
	group := []*Server{s1, s2, s3, s4, s5}

	startDone := make(chan int)
	go func(startDone chan int) {
		for i := 0; i < len(group); i++ {
			go group[i].Start()
			time.Sleep(time.Second)
			log.Printf("Server %s Started\n", group[i].Name)
			startDone <- 1
		}
	}(startDone)
	started := 0
	for i := range startDone {
		fmt.Println(i)
		started++
		if started == len(group) {
			log.Printf("All started")
			close(startDone)
			log.Printf("Start sending query")
			go func() {
				followers := []*Server{s2, s3, s4, s5}
				s1.InitLeader(followers)
				for i := 0; i < 10; i++ {
					log.Printf("Sending query %d", i)
					RepeatQuery(s1)
				}
			}()

		}
	}

	<-allDone

}

func RepeatQuery(leader *Server) {
	timeStamp := "Hello From" + time.Now().String()
	req := QueryArg{Command: timeStamp}
	if r, err := SendQueryRequest(leader, &req); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Receive reply from %s : %v", leader.Name, r)
	}
}

func SendQueryRequest(to *Server, req *QueryArg) (*QueryRes, error) {
	conn, err := grpc.Dial(to.Addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := NewRaftServiceClient(conn)
	return c.Query(context.Background(), req)
}
