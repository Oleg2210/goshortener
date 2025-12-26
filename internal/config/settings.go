package config

import (
	"flag"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

// минимальная длина id
const MinLength = 5

// максимальная длина id
const MaxLength = 10

var (
	PortAddres      string
	ResolveAddress  string
	FileStoragePath string
	DatabaseInfo    string
	AuthSecret      string
)

type envConfig struct {
	PortAddres      string `env:"SERVER_ADDRESS"`
	ResolveAddress  string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseInfo    string `env:"DATABASE_DSN"`
	AuthSecret      string `env:"AUTH_SECRET"`
}

func Load() {
	flag.StringVar(&PortAddres, "a", ":8080", "server address")
	flag.StringVar(&ResolveAddress, "b", "http://localhost:8080", "base URL")
	flag.StringVar(&FileStoragePath, "f", "urls-storage.json", "file storage")
	flag.StringVar(&DatabaseInfo, "d", "", "database dsn")
	flag.Parse()

	var e envConfig
	if err := cleanenv.ReadEnv(&e); err != nil {
		log.Fatalf("config error: %v", err)
	}

	if e.PortAddres != "" {
		PortAddres = e.PortAddres
	}
	if e.ResolveAddress != "" {
		ResolveAddress = e.ResolveAddress
	}
	if e.FileStoragePath != "" {
		FileStoragePath = e.FileStoragePath
	}
	if e.DatabaseInfo != "" {
		DatabaseInfo = e.DatabaseInfo
	}
	if e.AuthSecret == "" {
		AuthSecret = "secret"
	}
}
