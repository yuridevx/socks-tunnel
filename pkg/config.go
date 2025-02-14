package pkg

// Config holds the configuration values
type Config struct {
	LocalAddr  string `mapstructure:"local"`
	RemoteAddr string `mapstructure:"remote"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
}

var ConfigInst Config
