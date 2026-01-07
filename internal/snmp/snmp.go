package snmp

import (
	"fmt"

	"github.com/MustafaMertSandal/SNMP_Task/internal/config"
	"github.com/gosnmp/gosnmp"
)

type Client struct {
	snmp *gosnmp.GoSNMP
}

func New(cfg *config.Config) (*Client, error) {
	version := gosnmp.Version2c
	switch cfg.SNMP.Version {
	case "2c":
		version = gosnmp.Version2c
	default:
		return nil, fmt.Errorf("unsupported snmp.version: %s", cfg.SNMP.Version)
	}

	g := &gosnmp.GoSNMP{
		Target:         cfg.Target.Address,
		Port:           cfg.Target.Port,
		Community:      cfg.SNMP.Community,
		Version:        version,
		Timeout:        cfg.SNMP.Timeout.Duration,
		Retries:        cfg.SNMP.Retries,
		MaxRepetitions: cfg.SNMP.MaxRepetitions,
	}

	if err := g.Connect(); err != nil {
		return nil, err
	}
	return &Client{snmp: g}, nil
}

func (c *Client) Close() {
	if c.snmp != nil && c.snmp.Conn != nil {
		_ = c.snmp.Conn.Close()
	}
}

func (c *Client) Get(oids ...string) (*gosnmp.SnmpPacket, error) {
	return c.snmp.Get(oids)
}

func (c *Client) Walk(root string, fn gosnmp.WalkFunc) error {
	return c.snmp.BulkWalk(root, fn)
}
