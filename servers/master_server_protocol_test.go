package servers

import (
	"fmt"
	"testing"
)

const (
	masterServer string = "208.64.200.52:27011"
)

func TestGetServerList(t *testing.T) {
	serversChannel, errorChannel := GetServerList(masterServer, Region_RestOfTheWorld, `\gamedir\naturalselection2`)
	for {
		select {
		case <-serversChannel:
		case error := <-errorChannel:
			if error == ChannelExhausted {
				return
			} else {
				t.Fatalf("Error during server list fecthing: %s", error.Error())
			}
		}
	}
}

func BenchmarkGetServerList(b *testing.B) {
	serversChannel, errorChannel := GetServerList(masterServer, Region_RestOfTheWorld, `\gamedir\left4dead`)
	for {
		select {
		case <-serversChannel:
		case error := <-errorChannel:
			if error == ChannelExhausted {
				return
			} else {
				b.Fatalf("Error during server list fecthing: %s", error.Error())
			}
		}
	}
}

func ExampleGetServerList() {
	serversChannel, errorChannel := GetServerList(masterServer, Region_RestOfTheWorld, `\gamedir\naturalselection2`)
	for {
		select {
		case servers := <-serversChannel:
			// Handle this batch of servers. For example: allServers = append(allServers, servers)
			fmt.Printf("Fetched %d servers", len(servers))
		case error := <-errorChannel:
			// If an error occurs, it will be published on the error channel
			if error == ChannelExhausted {
				return
			} else {
				fmt.Errorf("Error during server fetch: %s", error.Error())
			}
		}
	}
}
