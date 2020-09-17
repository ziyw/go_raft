// Handling read and write from file.

package main

import (
	"bufio"
	_ "fmt"
	"log"
	"os"
	_ "strconv"
	_ "strings"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func AppendLine(file string, content string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer f.Close()
	_, err = f.Write([]byte(content))
	check(err)
}

func WriteLine(file string, content string) error {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(content)); err != nil {
		return err
	}
	return nil
}

func ReadLines(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return make([]string, 0), nil
	}
	defer f.Close()

	// limit of lines, not allow big files
	result := make([]string, 0, 100)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		result = append(result, string(scanner.Bytes()))
	}

	err = scanner.Err()
	return result, nil
}

// TODO: parse string to Entries
// TODO: parse string to int  strconv.ParseInt(s, 10, 64)
// TODO: parse string to antying
