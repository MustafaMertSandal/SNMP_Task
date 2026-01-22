package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	dd, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = dd
	return nil
}

type Config struct {
	SNMP struct {
		Version        string   `yaml:"version"`
		Community      string   `yaml:"community"`
		Timeout        Duration `yaml:"timeout"`
		Retries        int      `yaml:"retries"`
		MaxRepetitions uint32   `yaml:"max_repetitions"`
	} `yaml:"snmp"`

	Collect struct {
		System     bool `yaml:"system"`
		Interfaces bool `yaml:"interfaces"`
		IPRoutes   bool `yaml:"ip_routes"`
	} `yaml:"collect"`

	Targets []TargetConfig `yaml:"targets"`

	OIDs OIDs `yaml:"oids"`

	Database DatabaseConfig `yaml:"database"`

	Web WebConfig `yaml:"web"`
}

type WebConfig struct {
	Enabled *bool  `yaml:"enabled"`
	Addr    string `yaml:"addr"` // ":8080"
}

type TargetConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Port    uint16 `yaml:"port"`
}

type OIDs struct {
	System     SystemOIDs     `yaml:"system"`
	Interfaces InterfacesOIDs `yaml:"interfaces"`
	IPRoutes   IPRoutesOIDs   `yaml:"ip_routes"`
}

type SystemOIDs struct {
	SysDescr  string `yaml:"sysDescr"`
	SysUpTime string `yaml:"sysUpTime"`
	SysName   string `yaml:"sysName"`
}

type InterfacesOIDs struct {
	IfDescr      string `yaml:"ifDescr"`
	IfOperStatus string `yaml:"ifOperStatus"`
	IfInOctets   string `yaml:"ifInOctets"`
	IfOutOctets  string `yaml:"ifOutOctets"`
}

type IPRoutesOIDs struct {
	IpRouteDest    string `yaml:"ipRouteDest"`
	IpRouteIfIndex string `yaml:"ipRouteIfIndex"`
	IpRouteNextHop string `yaml:"ipRouteNextHop"`
	IpRouteType    string `yaml:"ipRouteType"`
	IpRouteMask    string `yaml:"ipRouteMask"`
}

type DatabaseConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"sslmode"`

	BatchSize int `yaml:"batch_size"`

	MaxConns        int32    `yaml:"max_conns"`
	MinConns        int32    `yaml:"min_conns"`
	MaxConnLifetime Duration `yaml:"max_conn_lifetime"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	if c.SNMP.Version == "" {
		return nil, fmt.Errorf("snmp.version required")
	}
	if c.SNMP.Version == "2c" && c.SNMP.Community == "" {
		return nil, fmt.Errorf("snmp.community required for v2c")
	}
	if c.SNMP.Timeout.Duration == 0 {
		c.SNMP.Timeout.Duration = 2 * time.Second
	}
	if c.SNMP.MaxRepetitions == 0 {
		c.SNMP.MaxRepetitions = 25
	}

	//Target validation
	if len(c.Targets) == 0 {
		return nil, fmt.Errorf("no targets configured: set target: or targets:")
	}
	for i := range c.Targets {
		if c.Targets[i].Name == "" {
			return nil, fmt.Errorf("targets[%d].name required", i)
		}
		if c.Targets[i].Address == "" {
			return nil, fmt.Errorf("targets[%d].address required", i)
		}
		if c.Targets[i].Port == 0 {
			c.Targets[i].Port = 161
		}
	}

	// OID validation (sadece ilgili collect açıksa zorunlu tutuyoruz)
	if c.Collect.System {
		if c.OIDs.System.SysDescr == "" || c.OIDs.System.SysUpTime == "" || c.OIDs.System.SysName == "" {
			return nil, fmt.Errorf("oids.system.sysDescr/sysUpTime/sysName must be set when collect.system=true")
		}
	}
	if c.Collect.Interfaces {
		o := c.OIDs.Interfaces
		if o.IfDescr == "" || o.IfOperStatus == "" || o.IfInOctets == "" || o.IfOutOctets == "" {
			return nil, fmt.Errorf("oids.interfaces.* must be set when collect.interfaces=true")
		}
	}
	if c.Collect.IPRoutes {
		o := c.OIDs.IPRoutes
		if o.IpRouteDest == "" || o.IpRouteIfIndex == "" || o.IpRouteNextHop == "" || o.IpRouteType == "" || o.IpRouteMask == "" {
			return nil, fmt.Errorf("oids.ip_routes.* must be set when collect.ip_routes=true")
		}
	}

	// Database defaults + validation (optional)
	// If you only want to print to stdout, set database.enabled=false in config.yaml
	if c.Database.Enabled {
		if c.Database.Port == 0 {
			c.Database.Port = 5432
		}
		if c.Database.SSLMode == "" {
			c.Database.SSLMode = "disable"
		}
		if c.Database.BatchSize == 0 {
			c.Database.BatchSize = 200
		}
		if c.Database.MaxConnLifetime.Duration == 0 {
			c.Database.MaxConnLifetime.Duration = 30 * time.Minute
		}
		if c.Database.Host == "" || c.Database.DBName == "" || c.Database.User == "" {
			return nil, fmt.Errorf("database.host, database.name and database.user must be set when database.enabled=true")
		}
	}

	return &c, nil
}
