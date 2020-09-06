package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

type LogEntry struct {
	command string
	term    uint32
}

type ServerState struct {
	// persist state, update before respond to rpc
	currentTerm uint32
	votedFor    uint32
	log         []LogEntry

	// versatile state
	commitedIndex uint32
	lastApplied   uint32
}

type LeaderState struct {
	serverState ServerState
	// volatile state
	nextIndex  []uint32
	matchIndex []uint32
}

func main() {

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
