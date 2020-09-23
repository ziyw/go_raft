package main

import (
	_ "fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"sync"
)

func main() {

	//	testLog := []*Entry{
	//		&Entry{Term: 1, Command: "Hello"},
	//		&Entry{Term: 1, Command: "World"},
	//	}
	//
	//	leader := Server{
	//		Name:        "Leader",
	//		Addr:        "localhost:30001",
	//		State:       Leader,
	//		currentTerm: 1,
	//		votedFor:    -1,
	//		log:         testLog,
	//		commitIndex: len(testLog) - 1,
	//		lastApplied: len(testLog) - 1,
	//	}
	//
	//	f1 := Server{Name: "NodeOne", Addr: "localhost:30002"}
	//	f2 := Server{Name: "NodeTwo", Addr: "localhost:30003"}
	//	f3 := Server{Name: "NodeThree", Addr: "localhost:30004"}
	//	s := []Server{leader, f1, f2, f3}
	//

	done := make(chan struct{})

	// Stage 1: get servers up and going.
	// TODO: this might not be the best way to do this. Need to refine later.
	s1 := NewServer("s1", "localhost:30001", 1)
	s2 := NewServer("s2", "localhost:30002", 2)
	s3 := NewServer("s3", "localhost:30003", 3)
	s4 := NewServer("s4", "localhost:30004", 4)
	s5 := NewServer("s5", "localhost:30005", 5)
	group := []*Server{s1, s2, s3, s4, s5}

	st1 := make(chan int)
	go StageOne(group, st1)

	st2 := make(chan int)
	st3 := make(chan int)
	go func() {
		for {
			select {
			case <-st1:
				log.Println("### Stage 1 is done")
				log.Println("### Stage 2 Start: Send query request from normal client. ###")
				go StageTwo(group, st2)
			case <-st2:
				log.Println("### Stage 2 is done ###")
				log.Println("### Stage 3 Start: Set leader in a group.  ###")
				go StageThree(group, st3)
			case <-st3:
				if l, err := group[0].GetLeader(); err != nil {
					log.Fatal(err)
				} else {
					log.Printf("Leader is %s\n", l.Name)
				}
				log.Println("### Stage 3 done###")
			case <-done:
				return
			}
		}
	}()

	<-done

	//
	//
	//
	//	allDone := make(chan int)
	//	go func() {
	//
	//		startDone := make(chan int)
	//		appendDone := make(chan int)
	//		voteDone := make(chan int)
	//
	//		go startAll(s, startDone)
	//
	//		for {
	//			select {
	//			case <-startDone:
	//				go leader.SendAppendRequest(&[]Server{f1, f2, f3}, appendDone)
	//			case <-appendDone:
	//				go leader.Vote(&[]Server{f1, f2, f3}, voteDone)
	//			case <-voteDone:
	//				allDone <- 1
	//			}
	//		}
	//	}()
	//
	//	<-allDone
}

// Stage 1: Start all servers.
// TODO: this is not the best way to count runnign servers. Need change.
func StageOne(group []*Server, done chan int) {
	log.Println("### Stage 1 Start: Setting up servers ###")
	var wg sync.WaitGroup

	for i := 0; i < len(group); i++ {
		wg.Add(1)
		go func(index int) {
			s := group[index]
			log.Println("start server: " + s.Name)
			wg.Done()
			if err := s.Start(); err != nil {
				panic("Server Starting Error")
			}
		}(i)
	}
	wg.Wait()
	done <- 1
}

// Stage 2: normal clients send request to server.
func StageTwo(group []*Server, done chan int) {
	var wg sync.WaitGroup
	for i := 0; i < len(group); i++ {
		wg.Add(1)
		go func(index int) {
			req := QueryArg{Command: "Hello"}
			if r, err := SendQueryRequest(group[index], &req); err != nil {
				log.Fatal(err)
			} else {
				log.Printf("Receive reply from %s : %v", group[index].Name, r)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	done <- 1
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

// Stage 3: Leader State
func StageThree(group []*Server, done chan int) {
	for _, v := range group {
		v.group = group
	}
	group[0].State = Leader
	done <- 1
}

//func startAll(servers []Server, done chan int) {
//	count := make(chan int)
//	for i := 0; i < len(servers); i++ {
//		go servers[i].Start(count)
//	}
//
//	go func() {
//		running := 0
//		for {
//			select {
//			case <-count:
//				log.Printf("Done ONE")
//				running++
//				if running == len(servers) {
//					log.Printf("All server started\n")
//					done <- 1
//					return
//				}
//			}
//		}
//
//	}()
//}
