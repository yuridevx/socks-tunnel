package pkg

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"text/template"
)

type ServiceData struct {
	Username         string
	ExecutablePath   string
	WorkingDirectory string
	LimitCPU         string
	MemoryLimit      string
}

func createServiceTemplate() string {
	return `[Unit]
Description=SOCKS Tunnel Service
After=network.target

[Service]
User={{.Username}}
ExecStart={{.ExecutablePath}} run
WorkingDirectory={{.WorkingDirectory}}
LimitCPU={{.LimitCPU}}
MemoryLimit={{.MemoryLimit}}
Restart=on-failure

[Install]
WantedBy=multi-user.target
`
}

func InstallService(c *cli.Context) error {
	username := "socks-tunnel-user"
	limitCPU := "500m"    // 50% of a single CPU core
	memoryLimit := "256M" // 256 Megabytes

	_, err := user.Lookup(username)
	if err != nil {
		// User doesn't exist, create it
		cmd := exec.Command("useradd", "-m", username)
		err = cmd.Run()
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
	}

	serviceFile := createServiceTemplate()

	executablePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	tmpl, err := template.New("service").Parse(serviceFile)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	data := ServiceData{
		Username:         username,
		ExecutablePath:   executablePath,
		WorkingDirectory: filepath.Dir(executablePath),
		LimitCPU:         limitCPU,
		MemoryLimit:      memoryLimit,
	}

	servicePath := "/etc/systemd/system/socks-tunnel.service"
	file, err := os.Create(servicePath)
	if err != nil {
		log.Fatalf("Failed to create service file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
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
