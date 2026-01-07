package main

import (
	"flag"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/gosnmp/gosnmp"

	"github.com/MustafaMertSandal/SNMP_Task/internal/config"
	"github.com/MustafaMertSandal/SNMP_Task/internal/convert"
	"github.com/MustafaMertSandal/SNMP_Task/internal/snmp"
)

func main() {
	cfgPath := flag.String("config", "../config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	fmt.Printf("[Target]\n	Name: %s\n	Address: %s:%d\n", cfg.Target.Name, cfg.Target.Address, cfg.Target.Port)

	client, err := snmp.New(cfg)
	if err != nil {
		log.Fatalf("snmp connect error: %v", err)
	}
	defer client.Close()

	if cfg.Collect.System {
		printSystem(client, cfg.OIDs.System)
	}
	if cfg.Collect.Interfaces {
		printInterfaces(client, cfg.OIDs.Interfaces)
	}
	if cfg.Collect.IPRoutes {
		printIPRoutes(client, cfg.OIDs.IPRoutes)
	}
	if cfg.Collect.UDPListeners {
		printUDPListeners(client, cfg.OIDs.UDP)
	}

	client.Close()
}

/* ---------------- System ---------------- */

func printSystem(client *snmp.Client, oids config.SystemOIDs) {
	pkt, err := client.Get(oids.SysDescr, oids.SysUpTime, oids.SysName)
	if err != nil {
		log.Printf("[System] get error: %v", err)
		return
	}

	var descr, name string
	var uptime uint64

	for _, vb := range pkt.Variables {
		switch strings.Trim(vb.Name, ".") {
		case oids.SysDescr:
			descr = convert.PDUToString(vb)
		case oids.SysUpTime:
			uptime = convert.PDUToUint64(vb)
		case oids.SysName:
			name = convert.PDUToString(vb)
		}
	}

	fmt.Println("\n[System]")
	fmt.Printf("	sysName   : %s\n", name)
	fmt.Printf("	sysUpTime : %d (TimeTicks)\n", uptime)
	fmt.Printf("	sysDescr  : %s\n", descr)
}

/* ---------------- Interfaces ---------------- */

type IfRow struct {
	Index      int
	Descr      string
	OperStatus int
	InOctets   uint64
	OutOctets  uint64
}

func printInterfaces(client *snmp.Client, oids config.InterfacesOIDs) {
	rows := map[int]*IfRow{}

	getRow := func(index int) *IfRow {
		r, ok := rows[index]
		if !ok {
			r = &IfRow{Index: index}
			rows[index] = r
		}
		return r
	}

	if err := client.Walk(oids.IfDescr, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).Descr = convert.PDUToString(pdu)
		return nil
	}); err != nil {
		log.Printf("[Interfaces] walk ifDescr error: %v", err)
		return
	}

	if err := client.Walk(oids.IfOperStatus, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).OperStatus = convert.PDUToInt(pdu)
		return nil
	}); err != nil {
		log.Printf("[Interfaces] walk ifOperStatus error: %v", err)
		return
	}

	if err := client.Walk(oids.IfInOctets, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).InOctets = convert.PDUToUint64(pdu)
		return nil
	}); err != nil {
		log.Printf("[Interfaces] walk ifInOctets error: %v", err)
		return
	}

	if err := client.Walk(oids.IfOutOctets, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).OutOctets = convert.PDUToUint64(pdu)
		return nil
	}); err != nil {
		log.Printf("[Interfaces] walk ifOutOctets error: %v", err)
		return
	}

	out := make([]IfRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })

	fmt.Println("\n[Interfaces]")
	for _, r := range out {
		fmt.Printf("	ifIndex=%d descr=%q oper=%s inOctets=%d outOctets=%d\n",
			r.Index, r.Descr, convert.OperStatusText(r.OperStatus), r.InOctets, r.OutOctets)
	}
}

/* ---------------- IP Routes ---------------- */

type RouteRow struct {
	Dest    string
	Mask    string
	NextHop string
	IfIndex int
	Type    int
}

func printIPRoutes(client *snmp.Client, oids config.IPRoutesOIDs) {
	rows := map[string]*RouteRow{}
	getRow := func(dest string) *RouteRow {
		r, ok := rows[dest]
		if !ok {
			r = &RouteRow{Dest: dest}
			rows[dest] = r
		}
		return r
	}

	_ = client.Walk(oids.IpRouteDest, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest)
		return nil
	})

	_ = client.Walk(oids.IpRouteMask, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest).Mask = convert.PDUToIPv4(pdu)
		return nil
	})

	_ = client.Walk(oids.IpRouteNextHop, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest).NextHop = convert.PDUToIPv4(pdu)
		return nil
	})

	_ = client.Walk(oids.IpRouteIfIndex, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest).IfIndex = convert.PDUToInt(pdu)
		return nil
	})

	_ = client.Walk(oids.IpRouteType, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest).Type = convert.PDUToInt(pdu)
		return nil
	})

	out := make([]RouteRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dest < out[j].Dest })

	fmt.Println("\n[IP Routes]")
	for _, r := range out {
		fmt.Printf("  dest=%s mask=%s nextHop=%s ifIndex=%d type=%s\n",
			r.Dest, r.Mask, r.NextHop, r.IfIndex, convert.RouteTypeText(r.Type))
	}
}

/* ---------------- UDP ---------------- */

type UDPListener struct {
	LocalAddr string
	LocalPort int
}

func printUDPListeners(client *snmp.Client, oids config.UDPOIDs) {
	rows := map[string]*UDPListener{}

	_ = client.Walk(oids.UdpLocalPort, func(pdu gosnmp.SnmpPDU) error {
		ints, err := convert.ParseLastNInts(pdu.Name, 5)
		if err != nil {
			return err
		}
		la := fmt.Sprintf("%d.%d.%d.%d", ints[0], ints[1], ints[2], ints[3])
		lp := ints[4]
		key := fmt.Sprintf("%s:%d", la, lp)
		rows[key] = &UDPListener{LocalAddr: la, LocalPort: lp}
		return nil
	})

	out := make([]*UDPListener, 0, len(rows))
	for _, v := range rows {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].LocalAddr == out[j].LocalAddr {
			return out[i].LocalPort < out[j].LocalPort
		}
		return out[i].LocalAddr < out[j].LocalAddr
	})

	fmt.Println("\n[UDP Listeners]")
	for _, u := range out {
		fmt.Printf("  %s:%d\n", u.LocalAddr, u.LocalPort)
	}
}
