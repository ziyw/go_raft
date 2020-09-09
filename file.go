// TODO: need to write simple test cases for it.

// Handling read and write from file.

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func ReadInt(file string) int64 {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	s := strings.TrimSpace(string(content))
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func WriteInt(file string, content int) {
	message := []byte(fmt.Sprintf("%d\n", content))
	err := ioutil.WriteFile(file, message, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
