package multibully

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const msgBlockSize = 128

const (
	// ElectionMessage is sent to announce an election
	ElectionMessage = iota
	// OKMessage is sent to confirm activity in an election
	OKMessage
	// CoordinatorMessage is sent to tell all the other nodes to followe it
	CoordinatorMessage
)

// Message is converted to and from bytes and sent between nodes
type Message struct {
	Kind uint8
	PID  uint64
	IP   *net.IP
}

// NewMessageFromBytes creates a new Message from the transmitted bytes
func NewMessageFromBytes(bytes []byte) *Message {
	data := string(bytes)
	tokens := strings.Split(data, "|")
	kind, _ := strconv.ParseUint(tokens[0], 10, 8)
	ip := net.ParseIP(tokens[1])
	pid, _ := strconv.ParseUint(tokens[2], 10, 64)

	return &Message{Kind: uint8(kind), IP: &ip, PID: uint64(pid)}
}

// Pack converts a Message into bytes for transmission
func (m *Message) Pack() []byte {
	ipString := net.IP.String(*m.IP)
	transmitData := fmt.Sprintf("%d|%s|%d|", m.Kind, ipString, m.PID)
	transmitData = transmitData + strings.Repeat("#", msgBlockSize-len(transmitData))
	return []byte(transmitData)
}
