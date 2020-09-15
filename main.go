package main

import (
	"fmt"
	"log"
)

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	ch := make(chan int)
	fmt.Println("Start Servers")

	leader := Server{Name: "Leader", Addr: "localhost:30001"}

	f1 := Server{Name: "NodeOne", Addr: "localhost:30002"}
	f2 := Server{Name: "NodeTwo", Addr: "localhost:30003"}
	f3 := Server{Name: "NodeThree", Addr: "localhost:30004"}

	go leader.Start()
	go f1.Start()
	go f2.Start()
	go f3.Start()

	req := AppendArg{
		Term:         1,
		LeaderId:     1,
		PrevLogIndex: 0,
		PrevLogTerm:  0,
		LeaderCommit: 0,
		Entries:      []*Entry{&Entry{Term: 1, Command: "Hello"}, &Entry{Term: 1, Command: "World"}},
	}

	fmt.Printf("Send request %v", req)
	r, err := leader.SendAppendRequest(f1, req)
	Check(err)
	fmt.Printf("Receive reponse %v", r)

	// 	followers := []Server{f1, f2, f3}
	//
	// 	for _, n := range followers {
	// 		req := AppendArg{}
	// 		leader.SendAppendRequest(n, req)
	// 	}

	ch <- 1

	// TODO: setup normal running situation
	// TODO: setup vote for leader situation
	// Normal running situation: setup one leader, 3 follower,
	// Leader send AppendEntryRequest to 3 follower to update log and as heartbeat

	<-ch
}
