package servers

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

// Defines a server region.
type ServerRegion byte

const (
	Region_USEastCoast    ServerRegion = 0x00
	Region_USWestCoast    ServerRegion = 0x01
	Region_SouthAmerica   ServerRegion = 0x02
	Region_Europe         ServerRegion = 0x03
	Region_Asia           ServerRegion = 0x04
	Region_Australia      ServerRegion = 0x05
	Region_MiddleEast     ServerRegion = 0x06
	Region_Africa         ServerRegion = 0x07
	Region_RestOfTheWorld ServerRegion = 0xFF

	cursor_FirstRequest string = "0.0.0.0:0"
)

// A server as returned by the master server.
type Server struct {
	firstOctet, secondOctet, thirdOctet, fourthOctet, port uint16
}

// Determine whether or not a Server is a null server (indicating end of server list) or not.
func (server Server) isNullServer() bool {
	return server.firstOctet == 0 && server.secondOctet == 0 && server.thirdOctet == 0 && server.fourthOctet == 0 && server.port == 0
}

// Returns the server's address in <ip>:<port> format.
func (server Server) String() string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", server.firstOctet, server.secondOctet, server.thirdOctet, server.fourthOctet, server.port)
}

type serverListQuery struct {
	messageType byte
	regionCode  byte
	cursor      string
	filter      string
}

func (query serverListQuery) bytes() []byte {
	buffer := new(bytes.Buffer)
	buffer.WriteByte(query.messageType)
	buffer.WriteByte(query.regionCode)
	buffer.WriteString(query.cursor)
	buffer.WriteByte(0x00)
	buffer.WriteString(query.filter)
	buffer.WriteByte(0x00)
	return buffer.Bytes()
}

// Use the generator concurrency pattern to stream the parsed servers over an established channel.
//
// The servers channel will be used to stream each newly read batch of Server objects. The error channel will serve
// as a "control" channel to indicate that an error has occurred at some point during the operation. To indicate that 
// there are no more servers to be streamed, a ChannelExhausted error will be sent on the error channel.
//
// A timeout value is required. For example: 500ms. See time.ParseDuration for more information.
func GetServerList(masterServer string, region ServerRegion, filter string, timeout string) (<-chan []Server, <-chan error) {
	servers := make(chan []Server)
	error := make(chan error)

	go func() {
		// Establish an outbound connection to the master server
		outboundConnection, connectError := connect(masterServer)
		if connectError != nil {
			error <- connectError
			return
		}
		defer outboundConnection.Close()

		// Establish an inbound connection to receive the master server replies
		inboundConnection, listenError := listen(outboundConnection.LocalAddr().(*net.UDPAddr))
		if listenError != nil {
			error <- listenError
			return
		}
		defer inboundConnection.Close()

		deadlineError := setReadDeadline(inboundConnection, timeout)
		if deadlineError != nil {
			error <- deadlineError
			return
		}

		query := serverListQuery{0x31, byte(region), cursor_FirstRequest, filter}
		for {
			_, writeError := outboundConnection.Write(query.bytes())
			if writeError != nil {
				error <- writeError
				return
			}

			reader := bufio.NewReader(inboundConnection)
			_, readServers, more, readError := readAndUnpack(reader)

			if readError != nil {
				error <- readError
				return
			}

			servers <- readServers

			if more {
				query = serverListQuery{0x31, byte(region), readServers[len(readServers)-1].String(), filter}
			} else {
				error <- ChannelExhausted
				return
			}
		}
	}()

	return servers, error
}

func readAndUnpack(reader *bufio.Reader) (header []byte, servers []Server, more bool, error error) {
	header, error = reader.ReadBytes(0x0A)
	if error != nil {
		return
	}

	more = true
	servers = make([]Server, 0)
	serverBuffer := make([]byte, 6)
	for reader.Buffered() >= 6 {
		if _, readError := reader.Read(serverBuffer); readError != nil {
			return nil, nil, false, readError
		}

		if server := unpackSingleServer(serverBuffer); !server.isNullServer() {
			servers = append(servers, server)
		} else {
			more = false
			break
		}
	}

	return
}

func unpackSingleServer(server []byte) Server {
	convertedServer := Server{uint16(server[0]), uint16(server[1]), uint16(server[2]), uint16(server[3]), binary.BigEndian.Uint16(server[4:])}
	return convertedServer
}
