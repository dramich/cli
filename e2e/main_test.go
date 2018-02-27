package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	cmd := setupTest()
	result := m.Run()
	teardownTest(cmd)
	os.Exit(result)
}

func setupTest() *exec.Cmd {
	source := os.Getenv("DAPPER_SOURCE")
	cmd := exec.Command(source+"/rancher", "--add-local", "--k8s-mode=internal")

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error startinga rancher server:%s", err)
		os.Exit(1)
	}

	var timeout int
	for {
		if timeout > 120 {
			fmt.Println("Timeout Reached")
			cmd.Process.Kill()
			os.Exit(1)
		}
		cmd2 := exec.Command("curl", "-k", "https://localhost:8443/ping")
		out, _ := cmd2.Output()
		//if err != nil{
		//	fmt.Printf("curl fail: %s", err)
		//	os.Exit(1)
		//}
		if strings.Contains(string(out), "pong") {
			break
		}
		timeout += 1
		time.Sleep(time.Second * 1)
	}
	return cmd
}

func teardownTest(cmd *exec.Cmd) {
	cmd.Process.Kill()
}
