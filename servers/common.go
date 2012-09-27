package servers

import (
	"errors"
	"net"
)

var (
	// Indicates that a channel has been exhausted.
	ChannelExhausted = errors.New("channel_exhausted")
)

// Connect to the given address (in <ip>:<port> form) and return the UDP connection or an error.
func connect(addr string) (connection *net.UDPConn, error error) {
	udpAddr, error := net.ResolveUDPAddr("udp", addr)
	if error != nil {
		return
	}

	connection, error = net.DialUDP("udp", nil, udpAddr)
	return
}

// Start listening on the given UDP address and return the UDP connection or an error.
func listen(addr *net.UDPAddr) (connection *net.UDPConn, error error) {
	connection, error = net.ListenUDP("udp", addr)
	return
}
