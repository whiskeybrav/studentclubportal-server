package configuration

import "github.com/BurntSushi/toml"

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Mail     MailConfig
}

type DatabaseConfig struct {
	Username string
	Password string
	Database string
}

type ServerConfig struct {
	Address string
	CORS    string
}

type MailConfig struct {
	FromAddress  string
	FromDisplay  string
	SendDomain   string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
}

func Configure() Config {
	config := Config{}
	_, err := toml.DecodeFile("config.toml", &config)
	if err != nil {
		panic(err)
	}
	return config
}
