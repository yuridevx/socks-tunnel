package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/proxy"
)

// Config holds the configuration values
type Config struct {
	LocalAddr  string `mapstructure:"localAddr"`
	RemoteAddr string `mapstructure:"remoteAddr"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
}

var config Config

var rootCmd = &cobra.Command{
	Use:   "socks-tunnel",
	Short: "A SOCKS5 tunnel application",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the SOCKS tunnel",
	Run:   runTunnel,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install as a systemd service",
	Run:   installService,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add flags to the run command
	runCmd.Flags().StringVar(&config.LocalAddr, "localAddr", "", "Local address to listen on")
	runCmd.Flags().StringVar(&config.RemoteAddr, "remoteAddr", "", "Remote SOCKS5 server address")
	runCmd.Flags().StringVar(&config.Username, "username", "", "Username for SOCKS5 authentication")
	runCmd.Flags().StringVar(&config.Password, "password", "", "Password for SOCKS5 authentication")

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(installCmd)

	// Bind environment variables
	viper.BindEnv("localAddr", "SOCKS_TUNNEL_LOCAL_ADDR")
	viper.BindEnv("remoteAddr", "SOCKS_TUNNEL_REMOTE_ADDR")
	viper.BindEnv("username", "SOCKS_TUNNEL_USERNAME")
	viper.BindEnv("password", "SOCKS_TUNNEL_PASSWORD")
}

func initConfig() {
	viper.SetConfigName("socks-tunnel") // name of config file (without extension)
	viper.SetConfigType("yaml")         // or viper.SetConfigType("json") if you use JSON

	// Add multiple config paths
	viper.AddConfigPath(".")                 // current directory
	viper.AddConfigPath("/etc/socks-tunnel") // system-wide config directory
	viper.AddConfigPath(os.Getenv("HOME"))   // user home directory

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	// Unmarshal the configuration into the Config struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}
}

func runTunnel(cmd *cobra.Command, args []string) {
	// Override config values with command line flags if they are set
	if localAddr, err := cmd.Flags().GetString("localAddr"); err == nil && localAddr != "" {
		config.LocalAddr = localAddr
	}
	if remoteAddr, err := cmd.Flags().GetString("remoteAddr"); err == nil && remoteAddr != "" {
		config.RemoteAddr = remoteAddr
	}
	if username, err := cmd.Flags().GetString("username"); err == nil && username != "" {
		config.Username = username
	}
	if password, err := cmd.Flags().GetString("password"); err == nil && password != "" {
		config.Password = password
	}

	listener, err := net.Listen("tcp", config.LocalAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", config.LocalAddr, err)
	}
	defer listener.Close()
	log.Printf("Listening on %s", config.LocalAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(clientConn, config.RemoteAddr, config.Username, config.Password)
	}
}

func handleConnection(clientConn net.Conn, remoteAddr, username, password string) {
	defer clientConn.Close()

	auth := proxy.Auth{
		User:     username,
		Password: password,
	}

	dialer, err := proxy.SOCKS5("tcp", remoteAddr, &auth, proxy.Direct)
	if err != nil {
		log.Printf("Failed to create SOCKS5 dialer: %v", err)
		return
	}

	remoteConn, err := dialer.Dial("tcp", remoteAddr)
	if err != nil {
		log.Printf("Failed to connect to remote SOCKS5 server: %v", err)
		return
	}
	defer remoteConn.Close()

	// Relay data between client and remote server
	go io.Copy(remoteConn, clientConn)
	io.Copy(clientConn, remoteConn)
}

func installService(cmd *cobra.Command, args []string) {
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
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
