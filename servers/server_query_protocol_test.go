package servers

import (
	"fmt"
	"testing"
)

func TestGetServerInfo(t *testing.T) {
	serverInfoChannel, errorChannel := GetServerInfo("81.19.212.190:27016")

	select {
	case serverInfo := <-serverInfoChannel:
		t.Logf("Received server info: %s", serverInfo)
	case error := <-errorChannel:
		t.Fatalf("Error during server list fecthing: %s", error.Error())
	}
}

func ExampleGetServerInfo() {
	serverInfoChannel, errorChannel := GetServerInfo("81.19.212.190:27016")

	select {
	case serverInfo := <-serverInfoChannel:
		fmt.Printf("Received server info: %s", serverInfo)
	case error := <-errorChannel:
		fmt.Errorf("Error during server list fecthing: %s", error.Error())
	}
}
