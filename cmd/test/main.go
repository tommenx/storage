package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/tommenx/storage/pkg/container"
	"github.com/tommenx/storage/pkg/isolate"
	"log"
	"time"
)

var (
	dockerId string
)

func init() {
	flag.StringVar(&dockerId, "docker", "", "identify docker")
}

func main() {
	flag.Parse()
	controller := container.NewClient()
	fmt.Printf("dockerId=%v", dockerId)
	cgroupPath, err := controller.GetCgroupPath(context.Background(), dockerId)
	if cgroupPath == "" {
		cgroupPath = "docker"
	}
	if err != nil {
		log.Printf("get container error, err=%+v", err)
		panic(err)
	}
	timer := time.NewTimer(5 * time.Second)
	speed := int64(10)
	for {
		select {
		case <-timer.C:
			update(cgroupPath, dockerId, speed)
			speed--
			timer.Reset(10 * time.Second)
		}
	}

}

func update(path string, dockerId string, speed int64) {
	if speed < 0 {
		speed = 1
	}
	request := map[string]int64{
		"write_bps_device": speed,
	}
	log.Printf("speed is %dMB", speed)
	err := isolate.SetBlkio(path, dockerId, request, "253", "5")
	if err != nil {
		log.Printf("set isolate error, err=%+v", err)
	}
}
