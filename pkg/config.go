package pkg

// Config holds the configuration values
type Config struct {
	LocalAddr  string `mapstructure:"localAddr"`
	RemoteAddr string `mapstructure:"remoteAddr"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
}

var ConfigInst Config
