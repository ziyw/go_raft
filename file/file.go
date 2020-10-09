package file

import (
	"bufio"
	"fmt"
	_ "io/ioutil"
	_ "log"
	"os"
	_ "strconv"
	_ "strings"
)

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

func AppendLines(file string, lines []string) error {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, l := range lines {
		_, err = f.Write([]byte(l))
	}
	return err
}

func DeleteFile(file string) error {
	return os.Remove(file)
}
