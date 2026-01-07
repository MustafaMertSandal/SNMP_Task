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

	Target struct {
		Name    string `yaml:"name"`
		Address string `yaml:"address"`
		Port    uint16 `yaml:"port"`
	} `yaml:"target"`

	Collect struct {
		System       bool `yaml:"system"`
		Interfaces   bool `yaml:"interfaces"`
		IPRoutes     bool `yaml:"ip_routes"`
		TCPConns     bool `yaml:"tcp_conns"`
		UDPListeners bool `yaml:"udp_listeners"`
	} `yaml:"collect"`

	OIDs OIDs `yaml:"oids"`
}

type OIDs struct {
	System     SystemOIDs     `yaml:"system"`
	Interfaces InterfacesOIDs `yaml:"interfaces"`
	IPRoutes   IPRoutesOIDs   `yaml:"ip_routes"`
	TCP        TCPOIDs        `yaml:"tcp"`
	UDP        UDPOIDs        `yaml:"udp"`
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

type TCPOIDs struct {
	TcpConnState string `yaml:"tcpConnState"`
}

type UDPOIDs struct {
	UdpLocalPort string `yaml:"udpLocalPort"`
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

	// Minimal validation + defaults
	if c.SNMP.Version == "" {
		return nil, fmt.Errorf("snmp.version required")
	}
	if c.SNMP.Version == "2c" && c.SNMP.Community == "" {
		return nil, fmt.Errorf("snmp.community required for v2c")
	}
	if c.Target.Address == "" {
		return nil, fmt.Errorf("target.address required")
	}
	if c.Target.Port == 0 {
		c.Target.Port = 161
	}
	if c.SNMP.Timeout.Duration == 0 {
		c.SNMP.Timeout.Duration = 2 * time.Second
	}
	if c.SNMP.MaxRepetitions == 0 {
		c.SNMP.MaxRepetitions = 25
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
	if c.Collect.UDPListeners {
		if c.OIDs.UDP.UdpLocalPort == "" {
			return nil, fmt.Errorf("oids.udp.udpLocalPort must be set when collect.udp_listeners=true")
		}
	}

	return &c, nil
}
