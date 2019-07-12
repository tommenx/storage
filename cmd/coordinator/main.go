package main

import "github.com/tommenx/storage/pkg/server"

func main() {
	s := server.NewServer()
	s.Run()
}
