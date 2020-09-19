package main

import (
	_ "fmt"
	"log"
)

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	done := make(chan int)
	leader := Server{Name: "Leader", Addr: "localhost:30001"}
	f1 := Server{Name: "NodeOne", Addr: "localhost:30002"}
	f2 := Server{Name: "NodeTwo", Addr: "localhost:30003"}
	f3 := Server{Name: "NodeThree", Addr: "localhost:30004"}

	start := make(chan string)

	go leader.Start(start)
	go f1.Start(start)
	go f2.Start(start)
	go f3.Start(start)

	count := 0
	select {
	case x := <-start:
		log.Printf("Build %s\n", x)
		count++
		log.Printf("Started %d\n", count)
		f := []Server{f1, f2, f3}
		if count == 3 {
			leader.SendAppendRequest(&f)
		}
		done <- 1
	default:
		log.Println("Waiting")
	}

	<-done

}
