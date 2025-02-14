package pkg

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
)

func InstallService(c *cli.Context) error {
	serviceFile := `[Unit]
Description=SOCKS Tunnel Service
After=network.target

[Service]
ExecStart=%s run
Restart=on-failure

[Install]
WantedBy=multi-user.target
`

	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	serviceContent := fmt.Sprintf(serviceFile, executablePath)
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
