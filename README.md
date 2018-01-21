# MultiBully

MultiBully is a Go library for distributed leadership election on a UDP multicast enabled network, like a LAN. It uses the [Bully algorithm](https://en.wikipedia.org/wiki/Bully_algorithm), a relatively simply mechanism which may not be suitable for a large numbers of nodes.

I think there are some bugs and inefficiencies in my implementation of it. It always manages to converge on the correct leader, but it sometimes takes a few goes round. Please feel free to help!

## Usage

The Bully algorithm elects the node with the largest `pid` to be leader. The `pid` doesn't necessarily need to be the process ID in the operating system – you could choose for this to be a timestamp or a fixed integer for each node. In this implementation, the combination of `pid` and IP address must be unique, so you can run multiple instances on a single host.

While MultiBully uses multicast UDP for communication between nodes, it transmits the non-multicast IP of each node in the communication packet, so you can point the follower at the new leader.

## Example

```go
package main

import (
	"github.com/tomtaylor/multibully"
	"log"
	"os"
	"net"
)

func main() {
	address := "224.0.0.0:9999"
	iface := "en0"
	stop := make(chan struct{})
	pid := uint64(os.Getpid())

	p, err := multibully.NewParticipant(address, iface, pid, func(state int, ip *net.IP) {
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
```

## Thanks

Thanks to [`oaStuff`](https://github.com/oaStuff/leaderElection) for their sample code which helped me understand how to implement the Bully algorithm.
