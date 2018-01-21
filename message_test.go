package multibully_test

import (
	"github.com/tomtaylor/multibully"
	"net"
	"testing"
)

func TestMessageToBytesAndBackIPv4(t *testing.T) {
	ip := net.ParseIP("192.168.1.1")
	pid := uint64(1)
	m1 := &multibully.Message{PID: pid, IP: &ip}
	bytes := m1.Pack()
	m2 := multibully.NewMessageFromBytes(bytes)

	if (m1.PID != m2.PID) || !m1.IP.Equal(*m2.IP) {
		t.Logf("Message 1: %+v", m1)
		t.Logf("Message 2: %+v", m2)
		t.Error("Message was not identical to and from bytes")
	}

	if len(bytes) > 128 {
		t.Error("Message was longer than expected 128 bytes")
	}
}

func TestMessageToBytesAndBackIPv6(t *testing.T) {
	ip := net.ParseIP("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	pid := uint64(1)
	m1 := &multibully.Message{PID: pid, IP: &ip}
	bytes := m1.Pack()
	m2 := multibully.NewMessageFromBytes(bytes)

	if (m1.PID != m2.PID) || !m1.IP.Equal(*m2.IP) {
		t.Logf("Message 1: %+v", m1)
		t.Logf("Message 2: %+v", m2)
		t.Error("Message was not identical to and from bytes")
	}

	if len(bytes) > 128 {
		t.Error("Message was longer than expected 128 bytes")
	}
}
