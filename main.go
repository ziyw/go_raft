package main

import (
	_ "fmt"
	"log"
	_ "sync"
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

	s1 := NewServer("s1", "localhost:30001", 1)
	s2 := NewServer("s2", "localhost:30002", 2)
	s3 := NewServer("s3", "localhost:30003", 3)
	s4 := NewServer("s4", "localhost:30004", 4)
	s5 := NewServer("s5", "localhost:30005", 5)

	doneStart := make(chan int)
	group := []*Server{s1, s2, s3, s4, s5}
	for i := 0; i < len(group); i++ {
		go func(index int) {
			s := group[index]
			log.Println("### Starting Server " + s.Name)
			if err := s.Start(); err != nil {
				panic("Server Starting Error")
			}
			doneStart <- 1
		}(i)
	}

	go func() {
		started := 0
		for {
			select {
			case <-doneStart:
				started++
				if started == len(group) {
					log.Println("Servers are started")
				}
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
