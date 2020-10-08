module go_raft

go 1.15

replace go_raft/raft => ./raft

require (
	github.com/golang/protobuf v1.4.2
	go_raft/raft v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20201006153459-a7d1128ccaa0
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)
