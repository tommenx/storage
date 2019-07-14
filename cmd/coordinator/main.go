package main

import (
	"github.com/tommenx/storage/pkg/server"
	"github.com/tommenx/storage/pkg/store"
)

func main() {
	db := store.NewEtcd([]string{"127.0.0.1:2379"})
	s := server.NewServer(db)
	s.Run()
}
