package main

import (
	_ "fmt"
	"log"
)

func main() {

	testLog := []*Entry{
		&Entry{Term: 1, Command: "Hello"},
		&Entry{Term: 1, Command: "World"},
	}

	leader := Server{
		Name:        "Leader",
		Addr:        "localhost:30001",
		IsLeader:    true,
		currentTerm: 1,
		votedFor:    -1,
		log:         testLog,
		commitIndex: len(testLog) - 1,
		lastApplied: len(testLog) - 1,
	}

	f1 := Server{Name: "NodeOne", Addr: "localhost:30002"}
	f2 := Server{Name: "NodeTwo", Addr: "localhost:30003"}
	f3 := Server{Name: "NodeThree", Addr: "localhost:30004"}
	s := []Server{leader, f1, f2, f3}

	allDone := make(chan int)
	go func() {

		startDone := make(chan int)
		appendDone := make(chan int)
		voteDone := make(chan int)

		go startAll(s, startDone)

		for {
			select {
			case <-startDone:
				go leader.SendAppendRequest(&[]Server{f1, f2, f3}, appendDone)
			case <-appendDone:
				go leader.Vote(&[]Server{f1, f2, f3}, voteDone)
			case <-voteDone:
				allDone <- 1
			}
		}
	}()

	<-allDone
}

func startAll(servers []Server, done chan int) {
	count := make(chan int)
	for i := 0; i < len(servers); i++ {
		go servers[i].Start(count)
	}

	go func() {
		running := 0
		for {
			select {
			case <-count:
				log.Printf("Done ONE")
				running++
				if running == len(servers) {
					log.Printf("All server started\n")
					done <- 1
					return
				}
			}
		}

	}()
}
