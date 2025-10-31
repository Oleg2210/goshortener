package config

import "flag"

const Letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const MinLength = 5
const MaxLength = 10

var (
	HostAddres     string
	ResolveAddress string
)

func ParseFlags() {
	flag.StringVar(&HostAddres, "a", ":8080", "server adress with port")
	flag.StringVar(&ResolveAddress, "b", "http://localhost:8080", "response URL")
	flag.Parse()
}
