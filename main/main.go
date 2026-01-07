package main

import (
	"flag"
	"fmt"
	"log"
	"sort"

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
	if cfg.Collect.TCPConns {
		printTCPConns(client, cfg.OIDs.TCP)
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
		switch vb.Name {
		case oids.SysDescr:
			descr = convert.PDUToString(vb)
		case oids.SysUpTime:
			uptime = convert.PDUToUint64(vb)
		case oids.SysName:
			name = convert.PDUToString(vb)
		}
	}

	fmt.Println("\n[System]")
	fmt.Printf("  sysName   : %s\n", name)
	fmt.Printf("  sysUpTime : %d (TimeTicks)\n", uptime)
	fmt.Printf("  sysDescr  : %s\n", descr)
}

/* ---------------- Interfaces ---------------- */

type IfRow struct {
	Index      int
	Descr      string
	OperStatus int
	InOctets   uint64
	OutOctets  uint64
}

func printInterfaces(c *snmp.Client, o config.InterfacesOIDs) {
	rows := map[int]*IfRow{}
	getRow := func(idx int) *IfRow {
		r, ok := rows[idx]
		if !ok {
			r = &IfRow{Index: idx}
			rows[idx] = r
		}
		return r
	}

	walk := func(root string, fn gosnmp.WalkFunc) error {
		return c.Walk(root, fn)
	}

	if err := walk(o.IfDescr, func(pdu gosnmp.SnmpPDU) error {
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

	_ = walk(o.IfOperStatus, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).OperStatus = convert.PDUToInt(pdu)
		return nil
	})

	_ = walk(o.IfInOctets, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).InOctets = convert.PDUToUint64(pdu)
		return nil
	})

	_ = walk(o.IfOutOctets, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).OutOctets = convert.PDUToUint64(pdu)
		return nil
	})

	out := make([]IfRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })

	fmt.Println("\n[Interfaces]")
	for _, r := range out {
		fmt.Printf("  ifIndex=%d descr=%q oper=%s inOctets=%d outOctets=%d\n",
			r.Index, r.Descr, operStatusText(r.OperStatus), r.InOctets, r.OutOctets)
	}
}

func operStatusText(v int) string {
	switch v {
	case 1:
		return "up"
	case 2:
		return "down"
	case 3:
		return "testing"
	case 4:
		return "unknown"
	case 5:
		return "dormant"
	case 6:
		return "notPresent"
	case 7:
		return "lowerLayerDown"
	default:
		return "n/a"
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

func printIPRoutes(c *snmp.Client, o config.IPRoutesOIDs) {
	rows := map[string]*RouteRow{}
	getRow := func(dest string) *RouteRow {
		r, ok := rows[dest]
		if !ok {
			r = &RouteRow{Dest: dest}
			rows[dest] = r
		}
		return r
	}

	_ = c.Walk(o.IpRouteDest, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest)
		return nil
	})

	_ = c.Walk(o.IpRouteMask, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest).Mask = convert.PDUToIPv4(pdu)
		return nil
	})

	_ = c.Walk(o.IpRouteNextHop, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest).NextHop = convert.PDUToIPv4(pdu)
		return nil
	})

	_ = c.Walk(o.IpRouteIfIndex, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest).IfIndex = convert.PDUToInt(pdu)
		return nil
	})

	_ = c.Walk(o.IpRouteType, func(pdu gosnmp.SnmpPDU) error {
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
			r.Dest, r.Mask, r.NextHop, r.IfIndex, routeTypeText(r.Type))
	}
}

func routeTypeText(v int) string {
	switch v {
	case 1:
		return "other"
	case 2:
		return "invalid"
	case 3:
		return "direct"
	case 4:
		return "indirect"
	default:
		return "n/a"
	}
}

/* ---------------- TCP ---------------- */

type TCPConn struct {
	LocalAddr string
	LocalPort int
	RemAddr   string
	RemPort   int
	State     int
}

func printTCPConns(c *snmp.Client, o config.TCPOIDs) {
	rows := map[string]*TCPConn{}

	_ = c.Walk(o.TcpConnState, func(pdu gosnmp.SnmpPDU) error {
		ints, err := convert.ParseLastNInts(pdu.Name, 10)
		if err != nil {
			return err
		}
		la := fmt.Sprintf("%d.%d.%d.%d", ints[0], ints[1], ints[2], ints[3])
		lp := ints[4]
		ra := fmt.Sprintf("%d.%d.%d.%d", ints[5], ints[6], ints[7], ints[8])
		rp := ints[9]
		key := fmt.Sprintf("%s:%d->%s:%d", la, lp, ra, rp)

		rows[key] = &TCPConn{
			LocalAddr: la, LocalPort: lp,
			RemAddr: ra, RemPort: rp,
			State: convert.PDUToInt(pdu),
		}
		return nil
	})

	out := make([]*TCPConn, 0, len(rows))
	for _, v := range rows {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].LocalAddr == out[j].LocalAddr {
			return out[i].LocalPort < out[j].LocalPort
		}
		return out[i].LocalAddr < out[j].LocalAddr
	})

	fmt.Println("\n[TCP Connections]")
	for _, t := range out {
		fmt.Printf("  %s:%d -> %s:%d state=%s\n",
			t.LocalAddr, t.LocalPort, t.RemAddr, t.RemPort, tcpStateText(t.State))
	}
}

func tcpStateText(v int) string {
	switch v {
	case 1:
		return "closed"
	case 2:
		return "listen"
	case 3:
		return "synSent"
	case 4:
		return "synReceived"
	case 5:
		return "established"
	case 6:
		return "finWait1"
	case 7:
		return "finWait2"
	case 8:
		return "closeWait"
	case 9:
		return "lastAck"
	case 10:
		return "closing"
	case 11:
		return "timeWait"
	case 12:
		return "deleteTCB"
	default:
		return "n/a"
	}
}

/* ---------------- UDP ---------------- */

type UDPListener struct {
	LocalAddr string
	LocalPort int
}

func printUDPListeners(c *snmp.Client, o config.UDPOIDs) {
	rows := map[string]*UDPListener{}

	_ = c.Walk(o.UdpLocalPort, func(pdu gosnmp.SnmpPDU) error {
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
