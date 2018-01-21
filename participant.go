package multibully

import (
	"bytes"
	"log"
	"net"
	"sync"
	"time"
)

const listenTimeout = 500
const leaderAnnouncementInterval = 250
const electionTimeout = 2000

// Callback is the function that is called when the node state changes
type Callback func(state int, leaderIP *net.IP)

const (
	// Follower is the state when we're following a Leader
	Follower = iota
	// Leader is the state when we're the primary node on the network
	Leader
)

type Participant struct {
	callback       Callback
	state          int
	transport      Transport
	IP             *net.IP
	pid            uint64
	txChan         chan *Message
	rxChan         chan *Message
	stop           bool
	waitGroup      *sync.WaitGroup
	electionTimer  *time.Timer
	announceTicker *time.Ticker
	listenTimer    *time.Timer
	leaderPid      uint64
}

func NewParticipant(address string, ifaceName string, pid uint64, callback Callback) (*Participant, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, err
	}

	sourceIP, err := getLocalInterfaceIPAddress(iface)
	if err != nil {
		return nil, err
	}

	ip := addr.IP
	port := addr.Port

	t, err := NewMulticastTransport(&ip, iface, port)
	if err != nil {
		return nil, err
	}

	return NewParticipantWithTransport(t, sourceIP, pid, callback), nil
}

func NewParticipantWithTransport(t Transport, IP *net.IP, pid uint64, callback Callback) *Participant {
	p := &Participant{
		callback:       callback,
		state:          Follower,
		transport:      t,
		IP:             IP,
		pid:            pid,
		txChan:         make(chan *Message),
		rxChan:         make(chan *Message),
		stop:           false,
		waitGroup:      &sync.WaitGroup{},
		electionTimer:  nil,
		announceTicker: nil,
		listenTimer:    nil,
		leaderPid:      0,
	}

	return p
}

// RunLoop is a blocking function that runs the Participant's connections
// You can stop by closing the blocking channel, done.
func (p *Participant) RunLoop(done chan struct{}) {
	// Tidy up loop
	go func() {
		<-done
		p.transport.Close()
	}()

	// Read loop
	go func() {
		p.waitGroup.Add(1)
		for {
			msg, err := p.transport.Read()
			if err != nil {
				log.Printf("< %s", err)
				break
			}

			if msg != nil {
				// Skip our own Messages, which we receive on the loopback interface
				if msg.PID == p.pid {
					continue
				}

				log.Printf("< %+v", msg)
				p.handleMessage(msg)
			}
		}
		p.waitGroup.Done()
	}()

	// Write loop
	go func() {
		p.waitGroup.Add(1)
	Loop:
		for {
			select {
			case msg := <-p.txChan:
				log.Printf("> %+v", msg)
				err := p.transport.Write(msg)
				if err != nil {
					log.Printf("> %s", err)
				}
			case <-done:
				break Loop
			}
		}
		p.waitGroup.Done()
	}()

	<-done
	p.waitGroup.Wait()
}

func (p *Participant) handleMessage(m *Message) {
	switch m.Kind {
	case ElectionMessage:
		p.handleElectionMessage(m)
	case OKMessage:
		p.handleOKMessage(m)
	case CoordinatorMessage:
		p.handleCoordinatorMessage(m)
	}
}

func (p *Participant) handleElectionMessage(m *Message) {
	if p.messageHasPriority(m) {
		p.stopAnnounceTicker()
	} else {
		p.sendMessage(OKMessage)
		p.StartElection()
	}
}

func (p *Participant) handleOKMessage(m *Message) {
	if p.messageHasPriority(m) {
		p.stopElection()
	}
}

func (p *Participant) handleCoordinatorMessage(m *Message) {
	if p.messageHasPriority(m) {
		p.becomeFollower(m.PID, m.IP)
		p.stopElection()
	} else {
		log.Print("* Received CoordinatorMessage from smaller PID, starting election")
		p.StartElection()
	}
}

func (p *Participant) becomeFollower(pid uint64, ip *net.IP) {
	if p.leaderPid != pid {
		p.state = Follower
		p.leaderPid = pid
		go p.callback(Follower, ip)
	}

	if pid == p.leaderPid {
		p.startListeningForLeader()
	}
}

func (p *Participant) StartElection() {
	p.stopAnnounceTicker()
	p.stopElection()
	p.sendMessage(ElectionMessage)

	p.electionTimer = time.AfterFunc(electionTimeout*time.Millisecond, func() {
		log.Println("* Nothing replied to election broadcast, become leader")
		p.becomeLeader()
	})
}

func (p *Participant) stopElection() {
	if p.electionTimer != nil {
		p.electionTimer.Stop()
	}
}

func (p *Participant) becomeLeader() {
	p.sendMessage(CoordinatorMessage)
	p.state = Leader
	p.leaderPid = p.pid
	go p.callback(Leader, nil)
	p.startAnnounceTicker()
}

func (p *Participant) startListeningForLeader() {
	p.stopListeningForLeader()
	p.listenTimer = time.AfterFunc(listenTimeout*time.Millisecond, func() {
		log.Println("* Leader did not broadcast within timeout, starting election")
		p.StartElection()
	})
}

func (p *Participant) stopListeningForLeader() {
	if p.listenTimer != nil {
		p.listenTimer.Stop()
	}
}

func (p *Participant) startAnnounceTicker() {
	p.stopAnnounceTicker()
	p.announceTicker = time.NewTicker(leaderAnnouncementInterval * time.Millisecond)

	go func() {
		for range p.announceTicker.C {
			p.sendMessage(CoordinatorMessage)
		}
	}()
}

func (p *Participant) stopAnnounceTicker() {
	if p.announceTicker != nil {
		p.announceTicker.Stop()
	}
}

func (p *Participant) sendMessage(kind uint8) {
	m := &Message{Kind: kind, PID: p.pid, IP: p.IP}
	p.txChan <- m
}

func (p *Participant) messageHasPriority(m *Message) bool {
	return (m.PID > p.pid) || (m.PID == p.pid && bytes.Compare(*m.IP, *p.IP) == 1)
}
