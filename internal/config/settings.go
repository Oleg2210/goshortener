package config

import "flag"

// символы, которые будут использоваться при генерации id
const Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// минимальная длина id
const MinLength = 5

// максимальная длина  id
const MaxLength = 10

var (
	PortAddres     string
	ResolveAddress string
)

func ParseFlags() {
	flag.StringVar(&PortAddres, "a", ":8080", "server adress with port")
	flag.StringVar(&ResolveAddress, "b", "http://localhost:8080", "response URL")
	flag.Parse()
}
