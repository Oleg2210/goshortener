package config

import (
	"flag"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	MinLength = 5
	MaxLength = 10

	CertHTTPS = "cert.pem"
	KeyHTTPS  = "key.pem"

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
	TrustedSubnet   string
)

func Load() error {
	flag.StringVar(&PortAddres, "a", "", "server address")
	flag.StringVar(&ResolveAddress, "b", "", "base URL")
	flag.StringVar(&FileStoragePath, "f", "", "file storage")
	flag.StringVar(&DatabaseInfo, "d", "", "database dsn")
	flag.StringVar(&AuditFile, "audit-file", "", "audit file path")
	flag.StringVar(&AuditURL, "audit-url", "", "audit url")
	flag.BoolVar(&EnableHTTPS, "s", false, "enable https")
	flag.StringVar(&ConfigFile, "c", "", "config file path")
	flag.StringVar(&TrustedSubnet, "t", "", "trusted subnet")

	flag.Parse()

	v := viper.New()

	v.SetDefault("server_address", DefaultPortAddres)
	v.SetDefault("base_url", DefaultResolveAddress)
	v.SetDefault("file_storage_path", DefaultFileStoragePath)
	v.SetDefault("auth_secret", "secret")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.BindEnv("server_address", "SERVER_ADDRESS")
	v.BindEnv("base_url", "BASE_URL")
	v.BindEnv("file_storage_path", "FILE_STORAGE_PATH")
	v.BindEnv("database_dsn", "DATABASE_DSN")
	v.BindEnv("auth_secret", "AUTH_SECRET")
	v.BindEnv("audit_file", "AUDIT_FILE")
	v.BindEnv("audit_url", "AUDIT_URL")
	v.BindEnv("enable_https", "ENABLE_HTTPS")
	v.BindEnv("trusted_subnet", "TRUSTED_SUBNET")
	v.BindEnv("config", "CONFIG")

	if ConfigFile == "" {
		ConfigFile = v.GetString("config")
	}

	if ConfigFile != "" {
		v.SetConfigFile(ConfigFile)

		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if PortAddres != "" {
		v.Set("server_address", PortAddres)
	}
	if ResolveAddress != "" {
		v.Set("base_url", ResolveAddress)
	}
	if FileStoragePath != "" {
		v.Set("file_storage_path", FileStoragePath)
	}
	if DatabaseInfo != "" {
		v.Set("database_dsn", DatabaseInfo)
	}
	if AuditFile != "" {
		v.Set("audit_file", AuditFile)
	}
	if AuditURL != "" {
		v.Set("audit_url", AuditURL)
	}
	if TrustedSubnet != "" {
		v.Set("trusted_subnet", TrustedSubnet)
	}
	if EnableHTTPS {
		v.Set("enable_https", true)
	}

	PortAddres = v.GetString("server_address")
	ResolveAddress = v.GetString("base_url")
	FileStoragePath = v.GetString("file_storage_path")
	DatabaseInfo = v.GetString("database_dsn")
	AuthSecret = v.GetString("auth_secret")
	AuditFile = v.GetString("audit_file")
	AuditURL = v.GetString("audit_url")
	EnableHTTPS = v.GetBool("enable_https")
	TrustedSubnet = v.GetString("trusted_subnet")

	return nil
}
