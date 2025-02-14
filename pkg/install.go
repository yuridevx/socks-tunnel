package pkg

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"os/user"
)

func InstallService(c *cli.Context) error {
	username := "socks-tunnel-user"

	_, err := user.Lookup(username)
	if err != nil {
		// User doesn't exist, create it
		cmd := exec.Command("useradd", "-m", username)
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	}

	serviceFile := `[Unit]
Description=SOCKS Tunnel Service
After=network.target

[Service]
User=%s
ExecStart=%s run
Restart=on-failure

[Install]
WantedBy=multi-user.target
`

	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	serviceContent := fmt.Sprintf(serviceFile, username, executablePath)
	servicePath := "/etc/systemd/system/socks-tunnel.service"

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		log.Fatalf("Failed to write service file: %v", err)
	}

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		log.Fatalf("Failed to reload systemd: %v", err)
	}

	if err := exec.Command("systemctl", "enable", "socks-tunnel").Run(); err != nil {
		log.Fatalf("Failed to enable service: %v", err)
	}

	log.Println("Service installed and enabled successfully.")
	return nil
}
