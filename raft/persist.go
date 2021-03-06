package raft

// Handle persist properties

import (
	"fmt"
	"github.com/ziyw/go_raft/file"
	"github.com/ziyw/go_raft/pb"
	"log"
	"os"
	"strconv"
	"strings"
)

func (s *Server) CurrentTerm() int {
	lines, err := file.ReadLines(s.termFile)
	if err != nil {
		return -1
	}
	t, err := strconv.ParseInt(strings.Trim(lines[0], "\n"), 10, 64)
	if err != nil {
		return -1
	}
	return int(t)
}

func (s *Server) SetCurrentTerm(t int) {
	f := s.termFile
	l := fmt.Sprintf("%d\n", t)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		file.AppendLines(f, []string{l})
		return
	}
	file.DeleteFile(f)
	file.AppendLines(f, []string{l})
}

func (s *Server) VotedFor() string {
	f := s.voteFile
	lines, err := file.ReadLines(f)
	if err != nil {
		return ""
	}
	return strings.Trim(lines[0], "\n ")
}

func (s *Server) SetVotedFor(v string) {
	f := s.voteFile
	l := fmt.Sprintf("%s\n", v)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		file.AppendLines(f, []string{l})
		return
	}
	file.DeleteFile(f)
	file.AppendLines(f, []string{l})
}

// TODO: Can be optimized
func (s *Server) Log() []*pb.Entry {
	f := s.logFile
	lines, err := file.ReadLines(f)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	entries := []*pb.Entry{}
	for _, l := range lines {
		p := strings.Split(l, ",")
		t, err := strconv.ParseInt(p[0], 10, 64)
		if err != nil {
			panic(err)
		}
		c := strings.Join(p[1:], ",")
		e := &pb.Entry{Term: t, Command: c}
		entries = append(entries, e)
	}
	return entries
}

func (s *Server) SetLog(entries []*pb.Entry) {
	lines := []string{}
	for _, e := range entries {
		l := fmt.Sprintf("%d,%s\n", e.Term, e.Command)
		lines = append(lines, l)
	}

	f := s.logFile
	file.DeleteFile(f)
	file.AppendLines(f, lines)
}

func (s *Server) AppendLog(e *pb.Entry) {
	l := fmt.Sprintf("%d,%s\n", e.Term, e.Command)
	file.AppendLines(s.logFile, []string{l})
}
