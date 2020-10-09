package raft

import (
	"bufio"
	"fmt"
	_ "io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func ReadInt(file string) (int, error) {
	last, _, err := LastLine(file)
	if err != nil {
		log.Fatal(err)
	}
	r, err := strconv.ParseInt(strings.Trim(last, "\n"), 10, 64)
	return int(r), err
}

func ReadLines(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		result = append(result, string(scanner.Bytes()))
	}
	err = scanner.Err()
	return result, err
}

func LastLine(file string) (string, int, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", -1, err
	}
	defer f.Close()

	count := 0
	result := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		result = string(scanner.Bytes())
		count++
	}
	err = scanner.Err()
	return result, count, err
}

func AppendLine(file string, line string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(line))
	return err
}

func ReadLine(file string, index int) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	count := 1
	result := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() && count < index {
		result = string(scanner.Bytes())
		count++
	}
	err = scanner.Err()
	if count > index {
		return "", fmt.Errorf("Read file out of range")
	}
	return result, err
}
