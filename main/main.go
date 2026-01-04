package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

const (
	oidSysName   = "1.3.6.1.2.1.1.5.0" // sysName.0
	oidSysUpTime = "1.3.6.1.2.1.1.3.0" // sysUpTime.0

	oidIfDescr       = "1.3.6.1.2.1.2.2.1.2"     // ifDescr.{ifIndex}
	oidIfOperStatus  = "1.3.6.1.2.1.2.2.1.8"     // ifOperStatus.{ifIndex}
	oidIfHCInOctets  = "1.3.6.1.2.1.31.1.1.1.6"  // ifHCInOctets.{ifIndex} (64-bit)
	oidIfHCOutOctets = "1.3.6.1.2.1.31.1.1.1.10" // ifHCOutOctets.{ifIndex} (64-bit)

	/*
			// System Group: 1.3.6.1.2.1.1
		oidSysDescr  = "1.3.6.1.2.1.1.1.0" // sysDescr.0
		oidSysUpTime = "1.3.6.1.2.1.1.3.0" // sysUpTime.0
		oidSysName   = "1.3.6.1.2.1.1.5.0" // sysName.0

		// ifTable: 1.3.6.1.2.1.2.2
		oidIfDescr      = "1.3.6.1.2.1.2.2.1.2"  // ifDescr.{ifIndex}
		oidIfOperStatus = "1.3.6.1.2.1.2.2.1.8"  // ifOperStatus.{ifIndex}
		oidIfInOctets   = "1.3.6.1.2.1.2.2.1.10" // ifInOctets.{ifIndex} (32-bit)
		oidIfOutOctets  = "1.3.6.1.2.1.2.2.1.16" // ifOutOctets.{ifIndex} (32-bit)
	*/
)

type IfRow struct {
	Index      int
	Descr      string
	OperStatus string
	InOctets   *big.Int
	OutOctets  *big.Int
}

func main() {
	target := flag.String("target", "192.168.1.1", "device IP/host")
	community := flag.String("community", "public", "SNMP v2c community")
	port := flag.Uint("port", 161, "SNMP port")
	timeout := flag.Duration("timeout", 2*time.Second, "SNMP timeout")
	retries := flag.Int("retries", 1, "SNMP retries")
	flag.Parse()

	g := &gosnmp.GoSNMP{
		Target:         *target,
		Port:           uint16(*port),
		Community:      *community,
		Version:        gosnmp.Version2c, // SNMP v2c
		Timeout:        *timeout,
		Retries:        *retries,
		MaxRepetitions: 25, // BulkWalk sırasında bir seferde kaç OID istenecek
	}

	if err := g.Connect(); err != nil {
		log.Fatalf("SNMP connect failed: %v", err)
	}
	defer g.Conn.Close()

	// 1) sysName + sysUpTime GET
	resp, err := g.Get([]string{oidSysName, oidSysUpTime})
	if err != nil {
		log.Fatalf("SNMP GET failed: %v", err)
	}

	sysName := pduToString(findPDU(resp.Variables, oidSysName))
	up := findPDU(resp.Variables, oidSysUpTime)
	fmt.Printf("sysName: %s\n", sysName)
	fmt.Printf("sysUpTime: %v (raw)\n\n", up.Value)

	// 2) ifTable BULK WALK: ifDescr / ifOperStatus / ifHCInOctets / ifHCOutOctets
	rows := map[int]*IfRow{}

	mustWalk := func(root string, fn func(pdu gosnmp.SnmpPDU) error) {
		err := g.BulkWalk(root, fn)
		if err != nil {
			log.Fatalf("BulkWalk %s failed: %v", root, err)
		}
	}

	// ifDescr
	mustWalk(oidIfDescr, func(pdu gosnmp.SnmpPDU) error {
		idx := parseIndex(pdu.Name)
		r := getRow(rows, idx)
		r.Descr = pduToString(pdu)
		return nil
	})

	// ifOperStatus
	mustWalk(oidIfOperStatus, func(pdu gosnmp.SnmpPDU) error {
		idx := parseIndex(pdu.Name)
		r := getRow(rows, idx)
		r.OperStatus = operStatusToText(pdu)
		return nil
	})

	// ifHCInOctets
	mustWalk(oidIfHCInOctets, func(pdu gosnmp.SnmpPDU) error {
		idx := parseIndex(pdu.Name)
		r := getRow(rows, idx)
		r.InOctets = pduToBigInt(pdu)
		return nil
	})

	// ifHCOutOctets
	mustWalk(oidIfHCOutOctets, func(pdu gosnmp.SnmpPDU) error {
		idx := parseIndex(pdu.Name)
		r := getRow(rows, idx)
		r.OutOctets = pduToBigInt(pdu)
		return nil
	})

	// çıktı güzel görünsün
	indexes := make([]int, 0, len(rows))
	for i := range rows {
		indexes = append(indexes, i)
	}
	sort.Ints(indexes)

	fmt.Println("Interfaces:")
	for _, i := range indexes {
		r := rows[i]
		// bazı cihazlarda HC sayaçlar yoksa nil gelebilir, o yüzden kontrol
		in := "<nil>"
		out := "<nil>"
		if r.InOctets != nil {
			in = r.InOctets.String()
		}
		if r.OutOctets != nil {
			out = r.OutOctets.String()
		}

		fmt.Printf("- ifIndex=%d descr=%q status=%s in=%s out=%s\n",
			r.Index, r.Descr, r.OperStatus, in, out)
	}
}

func getRow(m map[int]*IfRow, idx int) *IfRow {
	if m[idx] == nil {
		m[idx] = &IfRow{Index: idx}
	}
	return m[idx]
}

// OID adı "...<root>.<ifIndex>" şeklindedir. Sondaki sayıyı alıyoruz.
func parseIndex(oid string) int {
	parts := strings.Split(oid, ".")
	var idx int
	fmt.Sscanf(parts[len(parts)-1], "%d", &idx)
	return idx
}

func findPDU(vars []gosnmp.SnmpPDU, name string) gosnmp.SnmpPDU {
	for _, v := range vars {
		if v.Name == name {
			return v
		}
	}
	return gosnmp.SnmpPDU{}
}

func pduToString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", pdu.Value)
	}
}

// OperStatus integer döner: 1=up, 2=down, ...
func operStatusToText(pdu gosnmp.SnmpPDU) string {
	// pdu.Value genelde int/uint gibi gelir
	var n int
	switch v := pdu.Value.(type) {
	case int:
		n = v
	case uint:
		n = int(v)
	case int64:
		n = int(v)
	case uint64:
		n = int(v)
	default:
		return "unknown"
	}
	switch n {
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
		return "unknown"
	}
}

func pduToBigInt(pdu gosnmp.SnmpPDU) *big.Int {
	// Counter64 çoğunlukla uint64 olarak gelir
	switch v := pdu.Value.(type) {
	case uint64:
		return new(big.Int).SetUint64(v)
	case int64:
		if v < 0 {
			return big.NewInt(0)
		}
		return big.NewInt(v)
	case uint:
		return new(big.Int).SetUint64(uint64(v))
	case int:
		if v < 0 {
			return big.NewInt(0)
		}
		return big.NewInt(int64(v))
	default:
		// bazı cihazlarda farklı tip dönebilir
		return nil
	}
}
