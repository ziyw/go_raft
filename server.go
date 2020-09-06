package main

import (
	"golang.org/x/net/context"
)

type Server struct {
	Name string
	Addr string
}

func (s *Server) AppendArg(ctx context.Context, arg *AppendArg) (*AppendRes, error) {
	// TODO: handler the request
	return &AppendRes{Term: 1, Success: true}, nil
}
