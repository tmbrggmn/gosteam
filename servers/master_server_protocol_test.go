package servers

import (
	"fmt"
	"testing"
)

const (
	masterServer             string = "208.64.200.52:27011"
	unknownHostMasterServer  string = "this.domain.shouldnt.exist:27011"
	unresponsiveMasterServer string = "google.com:27011"
)

func TestGetServerList(t *testing.T) {
	serversChannel, errorChannel := GetServerList(masterServer, Region_RestOfTheWorld, `\gamedir\naturalselection2`, "500ms")
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

func TestGetServerList_InvalidTimeout(t *testing.T) {
	serversChannel, errorChannel := GetServerList(masterServer, Region_RestOfTheWorld, `\gamedir\naturalselection2`, "Horatio Longbottom")
	for {
		select {
		case <-serversChannel:
			t.Fatalf("This test is supposed to fail. It hasn't. Now go fix the timeout parsing function!")
		case error := <-errorChannel:
			t.Logf("Error: %s", error.Error())
			return
		}
	}
}

func TestGetServerList_UnknownHostMasterServer_CI(t *testing.T) {
	serversChannel, errorChannel := GetServerList(unknownHostMasterServer, Region_RestOfTheWorld, `\gamedir\naturalselection2`, "500ms")
	for {
		select {
		case <-serversChannel:
			t.Fatalf("This test expects an unknown hostname error but returned normally. It seems the hostname '%s' actually exists on this network.", unknownHostMasterServer)
		case error := <-errorChannel:
			t.Logf("Error: %s", error.Error())
			return
		}
	}
}

func TestGetServerList_UnresponsiveMasterServer(t *testing.T) {
	serversChannel, errorChannel := GetServerList(unresponsiveMasterServer, Region_RestOfTheWorld, `\gamedir\naturalselection2`, "1s")
	for {
		select {
		case <-serversChannel:
			t.Fatalf("This test expects no server list query response from '%s' but apparently it has in fact responded. Well, that's awkward.", unresponsiveMasterServer)
		case error := <-errorChannel:
			t.Logf("Error: %s", error.Error())
			return
		}
	}
}

func BenchmarkGetServerList(b *testing.B) {
	serversChannel, errorChannel := GetServerList(masterServer, Region_RestOfTheWorld, `\gamedir\left4dead`, "500ms")
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
	serversChannel, errorChannel := GetServerList(masterServer, Region_RestOfTheWorld, `\gamedir\naturalselection2`, "500ms")
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
