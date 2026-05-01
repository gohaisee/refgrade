package config

type Config struct {
	Port        string
	ServiceName string
}

type getenvFunc func(string) string

func LoadFromEnv(getenv getenvFunc) Config {
	return Config{
		Port:        getenv("PORT"),
		ServiceName: getenv("SERVICE_NAME"),
	}
}
