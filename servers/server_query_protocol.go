package servers

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	playerInfoHeaderType byte = 0x44
)

var (
	request_A2S_SERVER_INFO           []byte = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x54, 0x53, 0x6F, 0x75, 0x72, 0x63, 0x65, 0x20, 0x45, 0x6E, 0x67, 0x69, 0x6E, 0x65, 0x20, 0x51, 0x75, 0x65, 0x72, 0x79, 0x00}
	request_A2S_PLAYER_INFO_CHALLENGE []byte = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x55, 0xFF, 0xFF, 0xFF, 0xFF}
	request_A2S_PLAYER_INFO           []byte = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x55}
)

// The player information as returned by the game server.
type PlayerInfo struct {
	Type        byte
	PlayerCount int
	Players     []Player
}

func (playerInfo PlayerInfo) String() string {
	buffer := bytes.NewBufferString(fmt.Sprintf("%d players\n", playerInfo.PlayerCount))
	for _, player := range playerInfo.Players {
		buffer.WriteString(fmt.Sprintf("\t%s\n", player))
	}
	return buffer.String()
}

// Player-specific information tor each player on the game server.
type Player struct {
	Index  int
	Name   string
	Score  int32
	Uptime float32
}

func (player Player) String() string {
	return fmt.Sprintf("#%d %s (%d)", player.Index, player.Name, player.Score)
}

type challenge struct {
	header    byte
	challenge []byte
}

func GetPlayerInfo(server string, timeout string) (<-chan *PlayerInfo, <-chan error) {
	playerInfoChannel := make(chan *PlayerInfo)
	errorChannel := make(chan error)

	go func() {
		outboundConnection, inboundConnection, connectionError := openConnections(server, timeout)
		if connectionError != nil {
			errorChannel <- connectionError
			return
		}
		defer outboundConnection.Close()
		defer inboundConnection.Close()

		_, writeError := outboundConnection.Write(request_A2S_PLAYER_INFO_CHALLENGE)
		if writeError != nil {
			errorChannel <- writeError
			return
		}

		reader := bufio.NewReader(inboundConnection)

		// Read the challenge
		challengeBytes := make([]byte, 9)
		_, readError := reader.Read(challengeBytes)
		if readError != nil {
			errorChannel <- readError
			return
		}
		challenge := challenge{challengeBytes[4], challengeBytes[5:]}

		// Write the actual request including the previous challenge
		request := append(request_A2S_PLAYER_INFO, challenge.challenge...)
		_, writeError = outboundConnection.Write(request)
		if writeError != nil {
			errorChannel <- writeError
			return
		}

		// Peek at the header to verify that we got the expected reply
		headerBytes, peakError := reader.Peek(5)
		if peakError != nil {
			errorChannel <- peakError
			return
		}
		if headerBytes[4] != playerInfoHeaderType {
			errorChannel <- UnexpectedReply
		}

		// Read the player info
		buffer := new(bytes.Buffer)
		for reader.Buffered() > 0 {
			bytes := make([]byte, reader.Buffered())
			reader.Read(bytes)
			buffer.Write(bytes)
		}

		playerInfoChannel <- unpackPlayerInfo(buffer)
	}()

	return playerInfoChannel, errorChannel
}

func unpackPlayerInfo(buffer *bytes.Buffer) *PlayerInfo {
	playerInfo := new(PlayerInfo)
	_ = buffer.Next(4)
	playerInfo.Type, _ = buffer.ReadByte()
	playerInfo.PlayerCount = readNextByteAsInt(buffer)

	playerInfo.Players = make([]Player, playerInfo.PlayerCount)
	for i := 0; i < playerInfo.PlayerCount; i++ {
		player := new(Player)
		player.Index = readNextByteAsInt(buffer)
		player.Name = readNextString(buffer)
		player.Score = readNextLong(buffer)
		player.Uptime = readNextFloat(buffer)
		playerInfo.Players[i] = *player
	}

	return playerInfo
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
//
// A timeout value is required. For example: 500ms. See time.ParseDuration for more information.
func GetServerInfo(server string, timeout string) (<-chan *ServerInfo, <-chan error) {
	serverInfoChannel := make(chan *ServerInfo)
	errorChannel := make(chan error)

	go func() {
		outboundConnection, inboundConnection, connectionError := openConnections(server, timeout)
		if connectionError != nil {
			errorChannel <- connectionError
			return
		}
		defer outboundConnection.Close()
		defer inboundConnection.Close()

		_, writeError := outboundConnection.Write(request_A2S_SERVER_INFO)
		if writeError != nil {
			errorChannel <- writeError
			return
		}

		reader := bufio.NewReader(inboundConnection)
		headerBytes := make([]byte, 4)
		_, readError := reader.Read(headerBytes)
		if readError != nil {
			errorChannel <- readError
			return
		}

		buffer := bytes.NewBuffer(headerBytes)
		for reader.Buffered() > 0 {
			bytes := make([]byte, reader.Buffered())
			reader.Read(bytes)
			buffer.Write(bytes)
		}

		serverInfoChannel <- unpackServerInfo(buffer)
	}()

	return serverInfoChannel, errorChannel
}

func unpackServerInfo(buffer *bytes.Buffer) *ServerInfo {
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

func readNextLong(buffer *bytes.Buffer) int32 {
	var long int32
	_ = binary.Read(buffer, binary.LittleEndian, &long)
	return long
}

func readNextFloat(buffer *bytes.Buffer) float32 {
	var float float32
	_ = binary.Read(buffer, binary.LittleEndian, &float)
	return float
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

func openConnections(server string, timeout string) (outboundConnection *net.UDPConn, inboundConnection *net.UDPConn, error error) {
	outboundConnection, error = connect(server)
	if error != nil {
		return
	}

	inboundConnection, error = listen(outboundConnection.LocalAddr().(*net.UDPAddr))
	if error != nil {
		return
	}

	error = setReadDeadline(inboundConnection, timeout)
	if error != nil {
		return
	}

	return
}
