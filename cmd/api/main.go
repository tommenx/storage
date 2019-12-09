package main

import (
	"flag"
)

func init() {
	flag.Set("logtostderr", "true")

}

//
//func main() {
//	flag.Parse()
//	path := "/root/.kube/config"
//	cli, informer := controller.NewCliAndInformer(path)
//}
