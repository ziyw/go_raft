package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

// TODO: what need to be done for Leader
// Leader -> send aAppenEntires to all followers
//  update nextIndex and matchIndex from the response from follower

// func (s *Server) SendAppendRequest(other Server, req AppendArg) (*AppendRes, error) {
// 	conn, err := grpc.Dial(other.Addr, grpc.WithInsecure())
// 	Check(err)
// 	defer conn.Close()
//
// 	client := NewRaftServiceClient(conn)
// 	return client.AppendEntries(context.Background(), &req)
// }
//

func (s *Server) SendAppendRequest(servers *[]Server, done chan int) {
	log.Printf("Start Sending Append Request\n")

	followers := *servers

	s.nextIndex = make([]int, len(followers))
	s.matchIndex = make([]int, len(followers))
	for i := 0; i < len(followers); i++ {
		s.nextIndex[i] = len(s.log) - 1
		s.matchIndex[i] = 0
	}

	for i, f := range followers {
		log.Printf("Start Send Request to %s", f.Name)
		if len(s.log) < int(s.nextIndex[i]) {
			continue
		}

		conn, err := grpc.Dial(f.Addr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		client := NewRaftServiceClient(conn)

		request := AppendArg{
			Term:         s.currentTerm,
			PrevLogIndex: int64(s.nextIndex[i]),
			PrevLogTerm:  s.log[s.nextIndex[i]].Term,
			Entries:      s.log[s.nextIndex[i]:],
			LeaderCommit: int64(s.commitIndex),
		}
		reply, _ := client.AppendEntries(context.Background(), &request)

		log.Printf("Server %s Receive Reply %v\n", s.Name, reply)

		for {

			// Success repond
			if reply.Success {
				s.nextIndex[i] = s.commitIndex + 1
				s.matchIndex[i] = s.commitIndex
				break
			}

			s.nextIndex[i]--

			request := AppendArg{
				Term:         s.currentTerm,
				PrevLogIndex: int64(s.nextIndex[i]),
				PrevLogTerm:  s.log[s.nextIndex[i]].Term,
				Entries:      s.log[s.nextIndex[i]:],
				LeaderCommit: int64(s.commitIndex),
			}
			reply, _ = client.AppendEntries(context.Background(), &request)
			log.Printf("Server %s Receive Reply %v\n", s.Name, reply)
		}
	}
	done <- 1
}

func (s *Server) AppendEntries(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	log.Printf("Server %s Receive AppendEntries Request %v\n", s.Name, *arg)

	res := AppendRes{Term: s.currentTerm, Success: true}

	if arg.Term < s.currentTerm {
		res.Success = false
		return &res, nil
	}

	if len(s.log) == 0 {
		res.Success = true
		s.log = arg.Entries
		return &res, nil
	}

	if len(s.log) <= int(arg.PrevLogIndex) {
		res.Success = false
		return &res, nil
	}

	if s.log[arg.PrevLogIndex].Term != arg.PrevLogTerm {
		res.Success = false
		return &res, nil
	}

	// check conflit
	for i := 0; i < len(arg.Entries); i++ {
		j := int(i + int(arg.PrevLogIndex))
		if j >= len(s.log) {
			break
		}
		if s.log[j].Term != arg.Entries[i].Term {
			s.log = s.log[:j-1]
			break
		}
	}

	// Append entries that are not exsiting
	s.log = append(s.log, arg.Entries...)

	if int(arg.LeaderCommit) > s.commitIndex {
		s.commitIndex = min(int(arg.LeaderCommit), len(s.log)-1)
	}

	return &res, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
