package servers

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

var (
	request_A2S_INFO   []byte = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x54, 0x53, 0x6F, 0x75, 0x72, 0x63, 0x65, 0x20, 0x45, 0x6E, 0x67, 0x69, 0x6E, 0x65, 0x20, 0x51, 0x75, 0x65, 0x72, 0x79, 0x00}
	request_A2S_PLAYER []byte = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x55, 0xFF, 0xFF, 0xFF, 0xFF}
)

// The player information as returned by the game server.
type PlayerInfo struct {
	Type        byte
	PlayerCount int
	Players     []Player
}

// Player-specific information tor each player on the game server.
type Player struct {
	Index  int
	Name   string
	Score  int32
	Uptime float32
}

func (player Player) String() string {
	return fmt.Sprintf("%s (%d)", player.Name, player.Score)
}

// The server information as returned by the game server.
type ServerInfo struct {
	Header          []byte
	Type            byte
	ProtocolVersion byte
	Name            string
	Map             string
	GameDirectory   string
	GameDescription string
	GameVersion     string
	ApplicationID   int16
	NumberOfPlayers int
	MaximumPlayers  int
	NumberOfBots    int
	Password        bool
	VAC             bool
	Dedicated       string
	OperatingSystem string
	ExtraData       []byte
}

func (serverInfo ServerInfo) String() string {
	return fmt.Sprintf("%s (%d/%d on %s)", serverInfo.Name, serverInfo.NumberOfPlayers, serverInfo.MaximumPlayers, serverInfo.Map)
}

// Retrieve server information for the given server (in <ip>:<port> format).
//
// One of 2 things will happen: either the server information is published on the server information channel or an error is published
// on the error channel.
func GetServerInfo(server string) (<-chan *ServerInfo, <-chan error) {
	serverInfoChannel := make(chan *ServerInfo)
	errorChannel := make(chan error)

	go func() {
		outboundConnection, connectError := connect(server)
		if connectError != nil {
			errorChannel <- connectError
		}
		defer outboundConnection.Close()

		inboundConnection, listenError := listen(outboundConnection.LocalAddr().(*net.UDPAddr))
		if listenError != nil {
			errorChannel <- listenError
		}
		defer inboundConnection.Close()

		_, writeError := outboundConnection.Write(request_A2S_INFO)
		if writeError != nil {
			errorChannel <- writeError
		}

		reader := bufio.NewReader(inboundConnection)
		headerBytes := make([]byte, 4)
		_, readError := reader.Read(headerBytes)
		if readError != nil {
			errorChannel <- readError
		}

		buffer := bytes.NewBuffer(headerBytes)
		for reader.Buffered() > 0 {
			bytes := make([]byte, reader.Buffered())
			reader.Read(bytes)
			buffer.Write(bytes)
		}

		serverInfoChannel <- parseServerInfoReply(buffer)
	}()

	return serverInfoChannel, errorChannel
}

func parseServerInfoReply(buffer *bytes.Buffer) *ServerInfo {
	serverInfo := new(ServerInfo)
	serverInfo.Header = buffer.Next(4)
	serverInfo.Type, _ = buffer.ReadByte()
	serverInfo.ProtocolVersion, _ = buffer.ReadByte()
	serverInfo.Name = readNextString(buffer)
	serverInfo.Map = readNextString(buffer)
	serverInfo.GameDirectory = readNextString(buffer)
	serverInfo.GameDescription = readNextString(buffer)
	serverInfo.ApplicationID = readNextShort(buffer)
	serverInfo.NumberOfPlayers = readNextByteAsInt(buffer)
	serverInfo.MaximumPlayers = readNextByteAsInt(buffer)
	serverInfo.NumberOfBots = readNextByteAsInt(buffer)
	serverInfo.Dedicated = readNextByteAsString(buffer)
	serverInfo.OperatingSystem = readNextByteAsString(buffer)
	serverInfo.Password = readNextByteAsBool(buffer)
	serverInfo.VAC = readNextByteAsBool(buffer)
	serverInfo.GameVersion = readNextString(buffer)
	serverInfo.ExtraData = buffer.Next(buffer.Len())
	return serverInfo
}

func readNextString(buffer *bytes.Buffer) string {
	bytes, _ := buffer.ReadBytes(0x00)
	return string(bytes[0 : len(bytes)-1])
}

func readNextShort(buffer *bytes.Buffer) int16 {
	var short int16
	_ = binary.Read(buffer, binary.LittleEndian, &short)
	return short
}

func readNextByteAsInt(buffer *bytes.Buffer) int {
	byte, _ := buffer.ReadByte()
	return int(byte)
}

func readNextByteAsString(buffer *bytes.Buffer) string {
	byte, _ := buffer.ReadByte()
	return string(byte)
}

func readNextByteAsBool(buffer *bytes.Buffer) bool {
	switch byte, _ := buffer.ReadByte(); byte {
	case 0x01:
		return true
	}
	return false
}
