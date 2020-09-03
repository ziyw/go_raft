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
	currentTerm uint32
	votedFor    uint32
	log         []LogEntry
}

func persistState() {
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
