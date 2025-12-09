package config

import (
	"database/sql"
	"flag"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// минимальная длина id
const MinLength = 5

// максимальная длина  id
const MaxLength = 10

var (
	PortAddres      string
	ResolveAddress  string
	FileStoragePath string
	DatabaseInfo    string
	DB              *sql.DB
)

func ParseFlags() {
	flag.StringVar(&PortAddres, "a", ":8080", "server adress with port")
	flag.StringVar(&ResolveAddress, "b", "http://localhost:8080", "response URL")
	flag.StringVar(&FileStoragePath, "f", "urls-storage.json", "path to uls storage")
	flag.StringVar(&DatabaseInfo, "d", "", "database dsn connection")

	flag.Parse()

	if envPortAddres, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		PortAddres = envPortAddres
	}

	if envResolveAddress, ok := os.LookupEnv("BASE_URL"); ok {
		ResolveAddress = envResolveAddress
	}

	if fileStorage, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		FileStoragePath = fileStorage
	}

	if DatabaseDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		DatabaseInfo = DatabaseDSN
	}

	db, err := sql.Open("pgx", DatabaseInfo)

	if err == nil {
		DB = db
	}
}
