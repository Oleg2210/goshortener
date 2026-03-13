package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	// минимальная длина id
	MinLength = 5
	// максимальная длина id
	MaxLength              = 10
	CertHTTPS              = "cert.pem"
	KeyHTTPS               = "key.pem"
	DefaultPortAddres      = ":8080"
	DefaultResolveAddress  = "http://localhost:8080"
	DefaultFileStoragePath = "urls-storage.json"
)

var (
	PortAddres      string
	ResolveAddress  string
	FileStoragePath string
	DatabaseInfo    string
	AuthSecret      string
	AuditFile       string
	AuditURL        string
	EnableHTTPS     bool
	ConfigFile      string
)

type envConfig struct {
	PortAddres      string `env:"SERVER_ADDRESS"`
	ResolveAddress  string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseInfo    string `env:"DATABASE_DSN"`
	AuthSecret      string `env:"AUTH_SECRET"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	EnableHTTPS     bool   `env:"ENABLE_HTTPS"`
	ConfigFile      string `env:"CONFIG"`
}

type fileConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

func Load() error {
	flag.StringVar(&PortAddres, "a", "", "server address")
	flag.StringVar(&ResolveAddress, "b", "", "base URL")
	flag.StringVar(&FileStoragePath, "f", "", "file storage")
	flag.StringVar(&DatabaseInfo, "d", "", "database dsn")
	flag.StringVar(&AuditFile, "audit-file", "", "audit file path")
	flag.StringVar(&AuditURL, "audit-url", "", "audit url")
	flag.BoolVar(&EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&ConfigFile, "c", "", "audit url")

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
	if e.AuditFile != "" {
		AuditFile = e.AuditFile
	}
	if e.AuditURL != "" {
		AuditURL = e.AuditURL
	}
	if e.EnableHTTPS {
		EnableHTTPS = true
	}
	if e.ConfigFile != "" {
		ConfigFile = e.ConfigFile
	}

	if ConfigFile != "" {
		err := loadFileConfig()
		if err != nil {
			return err
		}
	}

	setDefaultValues()
	return nil
}

func loadFileConfig() error {
	file, err := os.Open(ConfigFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var fcfg fileConfig

	if err := json.NewDecoder(file).Decode(&fcfg); err != nil {
		return err
	}

	if fcfg.ServerAddress != "" {
		PortAddres = fcfg.ServerAddress
	}
	if fcfg.BaseURL != "" {
		ResolveAddress = fcfg.BaseURL
	}
	if fcfg.FileStoragePath != "" {
		FileStoragePath = fcfg.FileStoragePath
	}
	if fcfg.DatabaseDSN != "" {
		DatabaseInfo = fcfg.DatabaseDSN
	}
	if fcfg.EnableHTTPS {
		EnableHTTPS = true
	}

	return nil
}

func setDefaultValues() {
	if PortAddres == "" {
		PortAddres = DefaultPortAddres
	}
	if ResolveAddress == "" {
		ResolveAddress = DefaultPortAddres
	}
	if FileStoragePath == "" {
		FileStoragePath = DefaultFileStoragePath
	}
}
