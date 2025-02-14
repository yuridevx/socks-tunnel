package main

import (
	"github.com/spf13/viper"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

// Config holds the configuration values
type Config struct {
	LocalAddr  string `mapstructure:"localAddr"`
	RemoteAddr string `mapstructure:"remoteAddr"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
}

var config Config

func initConfig() {
	viper.SetConfigName("socks-tunnel") // name of config file (without extension)
	viper.SetConfigType("yaml")         // or viper.SetConfigType("json") if you use JSON

	// Add multiple config paths
	viper.AddConfigPath(".")                 // current directory
	viper.AddConfigPath("/etc/socks-tunnel") // system-wide config directory
	viper.AddConfigPath(os.Getenv("HOME"))   // user home directory

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file, %s", err)
	} else {
		// Unmarshal the configuration into the Config struct
		if err := viper.Unmarshal(&config); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
	}
}

func main() {
	initConfig()

	app := &cli.App{
		Name:  "socks-tunnel",
		Usage: "A SOCKS5 tunnel application",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "Run the SOCKS tunnel",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "localAddr",
						Usage:   "Local address to listen on",
						EnvVars: []string{"SOCKS_TUNNEL_LOCAL_ADDR"},
					},
					&cli.StringFlag{
						Name:    "remoteAddr",
						Usage:   "Remote SOCKS5 server address",
						EnvVars: []string{"SOCKS_TUNNEL_REMOTE_ADDR"},
					},
					&cli.StringFlag{
						Name:    "username",
						Usage:   "Username for SOCKS5 authentication",
						EnvVars: []string{"SOCKS_TUNNEL_USERNAME"},
					},
					&cli.StringFlag{
						Name:    "password",
						Usage:   "Password for SOCKS5 authentication",
						EnvVars: []string{"SOCKS_TUNNEL_PASSWORD"},
					},
				},
				Action: runTunnel,
			},
			{
				Name:   "install",
				Usage:  "Install as a systemd service",
				Action: installService,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
