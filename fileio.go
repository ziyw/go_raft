// Handling read and write from file.

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func PeekFile(file string) bool {
	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}

func SaveInt(file string, body int) error {
	s := []byte(fmt.Sprintf("%d\n", body))
	return ioutil.WriteFile(file, s, 0600)
}

func ReadInt(file string) (int, error) {
	s, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	r, err := strconv.ParseInt(strings.Trim(string(s), "\n"), 10, 64)
	return int(r), err
}

func SaveEntry(file string, entry *Entry) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	body := fmt.Sprintf("%d\t%s\n", entry.Term, entry.Command)
	if _, err := f.WriteString(body); err != nil {
		return err
	}
	return nil
}

func SaveEntries(file string, entries []*Entry) error {
	os.Remove(file)
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 0; i < len(entries); i++ {
		body := fmt.Sprintf("%d\t%s\n", entries[i].Term, entries[i].Command)
		if _, err := f.WriteString(body); err != nil {
			return err
		}
	}

	return nil
}

func ReadEntry(file string, index int) (*Entry, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		if count == index {
			l := string(scanner.Bytes())
			p := strings.Split(l, "\t")
			t, err := strconv.ParseInt(p[0], 10, 64)
			if err != nil {
				return nil, err
			}
			c := strings.Join(p[1:], "\t")
			e := &Entry{
				Term:    t,
				Command: c,
			}
			return e, nil
		}
		count++
		scanner.Bytes()
	}
	return nil, fmt.Errorf("Index out of range")
}

func ReadEntries(file string) ([]*Entry, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries := []*Entry{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := string(scanner.Bytes())
		p := strings.Split(l, "\t")
		t, err := strconv.ParseInt(p[0], 10, 64)
		if err != nil {
			return nil, err
		}
		c := strings.Join(p[1:], "\t")
		e := &Entry{
			Term:    t,
			Command: c,
		}
		entries = append(entries, e)
	}
	return entries, nil
}

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
