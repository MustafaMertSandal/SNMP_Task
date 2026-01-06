package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gosnmp/gosnmp"
)

func main() {
	g := &gosnmp.GoSNMP{
		Target:    "192.168.10.201",
		Port:      161,
		Community: "public",
		Version:   gosnmp.Version2c,
		Timeout:   2 * time.Second,
		Retries:   4,
		MaxOids:   gosnmp.MaxOids,
	}

	if err := g.Connect(); err != nil {
		log.Fatalf("SNMP connect failed: %v", err)
	}
	defer g.Conn.Close()

	// OID'ler
	//oidSysName := "1.3.6.1.2.1.1.5.0" // sysName.0
	//oidSysUp := "1.3.6.1.2.1.1.3.0"   // sysUpTime.0

	//***GET***
	/*resp, err := g.Get([]string{oidSysName, oidSysUp})
	if err != nil {
		log.Fatalf("SNMP get failed: %v", err)
	}

	for _, v := range resp.Variables {
		fmt.Printf("%s = %v\n", v.Name, v.Value)
	}*/

	//***BULKWALK***
	// Örn: ifDescr tablosu
	root := "1.3.6.1.2.1.2.2.1.2"

	err := g.BulkWalk(root, func(pdu gosnmp.SnmpPDU) error {
		if b, ok := pdu.Value.([]byte); ok {
			fmt.Printf("%s = %s\n", pdu.Name, string(b))
		} else {
			fmt.Printf("%s = %v\n", pdu.Name, pdu.Value)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("BulkWalk failed: %v", err)
	}

}
