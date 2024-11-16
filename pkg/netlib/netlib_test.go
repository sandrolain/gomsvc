package netlib

import (
	"fmt"
	"net"
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("GetFreePort returned error: %v", err)
	}
	if port <= 0 {
		t.Fatalf("GetFreePort returned an invalid port: %d", port)
	}

	// Check if the port is effectively free
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("The port %d is not free", port)
	}
	listener.Close()
}

func TestNetworkContainsIP(t *testing.T) {
	tests := []struct {
		network  string
		ip       string
		expected bool
	}{
		{
			network:  "192.168.0.0/24",
			ip:       "192.168.0.1",
			expected: true,
		},
		{
			network:  "192.168.0.0/24",
			ip:       "10.0.0.1",
			expected: false,
		},
	}

	for _, test := range tests {
		result, err := NetworkContainsIP(test.network, test.ip)
		if err != nil {
			t.Fatalf("NetworkContainsIP returned error: %v", err)
		}
		if result != test.expected {
			t.Fatalf("NetworkContainsIP(%q, %q) returned %v, expected %v", test.network, test.ip, result, test.expected)
		}
	}
}

func TestNetworkContainsIPError(t *testing.T) {
	_, err := NetworkContainsIP("", "192.168.0.1")
	if err == nil {
		t.Fatalf("NetworkContainsIP with empty network should return error")
	}
}
