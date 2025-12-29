package main

import (
	"fmt"

	"github.com/MustafaMertSandal/SNMP_Task/internal/config"
)

func main() {
	cfg, err := config.Load("../configs/config.yaml")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Config loaded from %s\n", cfg.Server.Addr)
}
