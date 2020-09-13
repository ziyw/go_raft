package main

func setup() {
	s1 := Server{Name: "NodeOne", Addr: "localhost:60001"}
	s2 := Server{Name: "NodeTwo", Addr: "localhost:60002"}

	s1.log = make([]Entry, 10)
	s2.log = make([]Entry, 10)

	go s1.Start()
	go s2.Start()

	s1.SendAppendRequest(s2)
}

func main() {
	// TODO: setup normal running situation
	// TODO: setup vote for leader situation
	// Normal running situation: setup one leader, 3 follower,
	// Leader send AppendEntryRequest to 3 follower to update log and as heartbeat
}
