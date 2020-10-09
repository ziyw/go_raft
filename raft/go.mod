module go_raft/raft

go 1.15

replace go_raft/pb => ../pb

replace go_raft/file => ../file

require (
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/stretchr/testify v1.6.1
	go_raft/pb v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20190311183353-d8887717615a
	google.golang.org/grpc v1.33.0
	google.golang.org/protobuf v1.25.0 // indirect
)
