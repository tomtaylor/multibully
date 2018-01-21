package multibully

import (
	"errors"
	"log"
	"net"
)

type Transport interface {
	Read() (*Message, error)
	Write(*Message) error
	Close() error
}

type MulticastTransport struct {
	readConn  *net.UDPConn
	writeConn *net.UDPConn
	buffer    []byte
}

func NewMulticastTransport(mcastIP *net.IP, mcastInterface *net.Interface, port int) (*MulticastTransport, error) {
	if !mcastIP.IsMulticast() {
		return nil, errors.New("Address supplied is not a multicast address")
	}

	listenIP := *mcastIP
	listenAddr := &net.UDPAddr{IP: listenIP, Port: port}
	log.Printf("* Listening on: %+v", listenAddr)

	readConn, err := net.ListenMulticastUDP("udp", mcastInterface, listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	broadcastAddr := &net.UDPAddr{IP: *mcastIP, Port: port}
	log.Printf("* Broadcasting on: %+v", broadcastAddr)
	writeConn, err := net.DialUDP("udp", nil, broadcastAddr)
	if err != nil {
		log.Fatal(err)
	}

	return &MulticastTransport{readConn: readConn, writeConn: writeConn, buffer: []byte{}}, nil
}

func (t *MulticastTransport) Read() (*Message, error) {
	readBuffer := make([]byte, 1500)
	var msg *Message
	var err error
Loop:
	for {
		num, _, e := t.readConn.ReadFrom(readBuffer)
		if err != nil {
			log.Println(err)
			err = e
		}

		t.buffer = append(t.buffer, readBuffer[:num]...)
		if len(t.buffer) >= msgBlockSize {
			data := t.buffer[:msgBlockSize]
			msg = NewMessageFromBytes(data)
			t.buffer = t.buffer[msgBlockSize:]
			break Loop
		}
	}

	return msg, err
}

func (t *MulticastTransport) Write(m *Message) error {
	bytes := m.Pack()
	_, err := t.writeConn.Write(bytes)
	return err
}

func (t *MulticastTransport) Close() error {
	if err := t.readConn.Close(); err != nil {
		return err
	}

	if err := t.writeConn.Close(); err != nil {
		return err
	}

	return nil
}

// TODO: this should handle IPv6 addresses
func getLocalInterfaceIPAddress(ifi *net.Interface) (*net.IP, error) {
	addrs, err := ifi.Addrs()
	if err != nil {
		return nil, err
	}

	for _, add := range addrs {
		if ipnet, ok := add.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return &ipnet.IP, nil
			}
		}
	}

	return nil, errors.New("No local interface address found")
}
