package main

import (
	"github.com/tomtaylor/multibully"
	"log"
	"os"
	"net"
	"flag"
)


func main() {
	var iface *string
	iface = flag.String("iface", "en0", "interface for multicast broadcast")
	flag.Parse()

	address := "224.0.0.0:9999"
	stop := make(chan struct{})
	pid := uint64(os.Getpid())

	p, err := multibully.NewParticipant(address, *iface, pid, func(state int, ip *net.IP) {
		switch state {
		case multibully.Follower:
			log.Println("* Became Follower of", ip)
		case multibully.Leader:
			log.Println("* Became Leader")
		}
	})

	if err != nil {
		log.Fatal(err)
	}

	go p.StartElection()
	p.RunLoop(stop)
}
