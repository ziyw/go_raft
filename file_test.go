package main

import (
	"fmt"
	"testing"
)

// unit test for file io
// TODO: need a better test cases for this

func TestReadSingleValueFromFile(t *testing.T) {}

func TestReadArrayFromFile(t *testing.T) {}

func TestWriteArrayToFile(t *testing.T) {}

func TestWriteEntryToFile(t *testing.T) {}

func TestReadEntryFromFile(t *testing.T) {}

func TestAppendLine(t *testing.T) {
	f := "test.txt"
	l1, l2 := "Line One", "Line Two"
	AppendLine(f, l1)
	AppendLine(f, l2)
	result := ReadLines(f)
	fmt.Printf("%#v", result)
}

func TestWriteLine(t *testing.T) {
	WriteLine("test_write.txt", "One Line")
	WriteLine("test_write.txt", "One Line")
	WriteLine("test_write.txt", "One Line")
}
