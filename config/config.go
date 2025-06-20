package config

type Config struct {
	RabbitMQ RabbitMQConfig
	SMTP     SMTPConfig
	API      APIConfig
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type APIConfig struct {
	Port string
}

func NewConfig() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig{
			Host:     "localhost",
			Port:     "5672",
			User:     "admin",
			Password: "admin",
		},
		SMTP: SMTPConfig{
			Host: "smtp.gmail.com",
			Port: 587,
			// These will be set via environment variables
			User:     "",
			Password: "",
			From:     "",
		},
		API: APIConfig{
			Port: "8080",
		},
	}
} 