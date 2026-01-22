package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/gosnmp/gosnmp"

	"github.com/MustafaMertSandal/SNMP_Task/frontend"
	"github.com/MustafaMertSandal/SNMP_Task/internal/config"
	"github.com/MustafaMertSandal/SNMP_Task/internal/convert"
	"github.com/MustafaMertSandal/SNMP_Task/internal/db"
	"github.com/MustafaMertSandal/SNMP_Task/internal/snmp"
)

func main() {

	// Config'deki degerler icin cfg tanimlama
	cfgPath := flag.String("config", "../config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	var store *db.Store

	// Database enable ise pool ve store'u tanimla.
	if cfg.Database.Enabled {

		//Pool tanimlama
		pool, err := db.NewPool(ctx, cfg.Database)
		if err != nil {
			log.Fatalf("db connect error: %v", err)
		}

		//Store tanimlama
		store = db.NewStore(pool, cfg.Database.BatchSize)
		defer store.Close()
	}

	// Her bir target için bir tane goroutine başlatalım.
	for _, v := range cfg.Targets {
		go pollLoop(ctx, 10*time.Second, cfg, v, store)
	}

	// UI + API server (default :8080)
	if _, err := frontend.StartWebServer(ctx, cfg, store); err != nil {
		log.Fatalf("web server start error: %v", err)
	}

	fmt.Println("Polling + Web UI started. Open http://localhost" + cfg.Web.Addr)

	fmt.Println("Polling started. Press Ctrl+C to stop.")
	<-sigCh // burada sen kapatana kadar bekler
	fmt.Println("\nStopping...")
	cancel()

	// Goroutine'in düzgün kapanması için küçük bekleme (opsiyonel)
	time.Sleep(300 * time.Millisecond)
}

/* ---------------- Goroutine ---------------- */

// *----------------- Poll --------------------*//
func pollOnce(ctx context.Context, cfg *config.Config, trgt config.TargetConfig, store *db.Store) error {

	// SNMP icin client tanimlama
	client, err := snmp.New(cfg, trgt)
	if err != nil {
		log.Fatalf("snmp connect error: %v", err)
	}
	defer client.Close()

	var deviceID int64

	// Database enable ise pool ve store'u tanimla.
	if cfg.Database.Enabled {

		//Device'i ekle ve id'sini al.
		deviceID, err = store.EnsureDevice(ctx, trgt.Name)
		if err != nil {
			log.Fatalf("db ensure device error: %v", err)
		}
	}

	pollTime := time.Now().UTC()

	// System enable ise system bilgilerini (sysName, sysUpTime ve sysDescr) snmp ile alir ve database'e ekler.
	if cfg.Collect.System {
		sysName, uptime, sysDescr, err := collectSystem(client, cfg.OIDs.System)

		if err != nil {
			log.Printf("[System] collect error: %v", err)
		} else {
			if store != nil {
				if err := store.InsertRouterSNMP(ctx, pollTime, deviceID, sysName, int64(uptime), sysDescr); err != nil {
					log.Printf("[DB] insert router_snmp error: %v", err)
				}
			}
		}
	}

	// Interfaces enable ise interface bilgilerini (ifIndex, descr, ifOperStatus, ifInOctets ve ifOutOctets) snmp ile alir ve database'e ekler.
	if cfg.Collect.Interfaces {
		ifRows, err := collectInterfaces(client, cfg.OIDs.Interfaces)

		if err != nil {
			log.Printf("[Interfaces] collect error: %v", err)
		} else {
			if store != nil {
				metrics := make([]db.IfMetric, 0, len(ifRows))
				for _, r := range ifRows {
					metrics = append(metrics, db.IfMetric{
						IfIndex:    r.Index,
						IfDescr:    r.Descr,
						OperStatus: r.OperStatus,
						InOctets:   r.InOctets,
						OutOctets:  r.OutOctets,
					})
				}
				if err := store.InsertInterfaceMetrics(ctx, pollTime, deviceID, metrics); err != nil {
					log.Printf("[DB] insert router_interface_metrics error: %v", err)
				}
			}
		}
	}

	// IPRoutes enable ise IPRoutes bilgilerini (dest, mask, nextHop, ifIndex ve type) snmp ile alir ve database'e ekler.
	if cfg.Collect.IPRoutes {
		routes, err := collectIPRoutes(client, cfg.OIDs.IPRoutes)
		if err != nil {
			log.Printf("[IP Routes] collect error: %v", err)
		} else {
			if store != nil {
				dbRoutes := make([]db.RouteRow, 0, len(routes))
				for _, r := range routes {
					// Avoid invalid inet casts
					if r.Dest == "" || r.Mask == "" || r.NextHop == "" {
						continue
					}
					dbRoutes = append(dbRoutes, db.RouteRow{
						Dest:    r.Dest,
						Mask:    r.Mask,
						NextHop: r.NextHop,
						IfIndex: r.IfIndex,
						Type:    r.Type,
					})
				}
				if err := store.InsertIPRoutes(ctx, pollTime, deviceID, dbRoutes); err != nil {
					log.Printf("[DB] insert router_ip_routes error: %v", err)
				}
			}
		}
	}
	client.Close()
	return nil
}

func pollLoop(ctx context.Context, wait time.Duration, cfg *config.Config, trgt config.TargetConfig, store *db.Store) {
	ticker := time.NewTicker(wait)
	defer ticker.Stop()

	//Program başlar başlamaz 1 kere poll:
	if err := pollOnce(ctx, cfg, trgt, store); err != nil {
		fmt.Println("poll error:", err)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("pollLoop stopped:", ctx.Err())
			return
		case <-ticker.C:
			if err := pollOnce(ctx, cfg, trgt, store); err != nil {
				fmt.Println("poll error:", err)
			}
		}
	}
}

/* ---------------- System ---------------- */

// System bilgilerini snmp ile alir ve dogru formata cevirir.
func collectSystem(client *snmp.Client, oids config.SystemOIDs) (sysName string, uptime uint64, sysDescr string, err error) {
	pkt, err := client.Get(oids.SysDescr, oids.SysUpTime, oids.SysName)
	if err != nil {
		return "", 0, "", err
	}

	for _, vb := range pkt.Variables {
		switch strings.Trim(vb.Name, ".") {
		case oids.SysDescr:
			sysDescr = convert.PDUToString(vb)
		case oids.SysUpTime:
			uptime = convert.PDUToUint64(vb)
		case oids.SysName:
			sysName = convert.PDUToString(vb)
		}
	}

	return sysName, uptime, sysDescr, nil
}

/* ---------------- Interfaces ---------------- */

type IfRow struct {
	Index      int
	Descr      string
	OperStatus int
	InOctets   uint64
	OutOctets  uint64
}

func collectInterfaces(client *snmp.Client, oids config.InterfacesOIDs) ([]IfRow, error) {
	rows := map[int]*IfRow{}

	getRow := func(index int) *IfRow {
		r, ok := rows[index]
		if !ok {
			r = &IfRow{Index: index}
			rows[index] = r
		}
		return r
	}

	//IfDescr
	if err := client.Walk(oids.IfDescr, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).Descr = convert.PDUToString(pdu)
		return nil
	}); err != nil {
		return nil, err
	}

	//IfOperStatus
	if err := client.Walk(oids.IfOperStatus, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).OperStatus = convert.PDUToInt(pdu)
		return nil
	}); err != nil {
		return nil, err
	}

	//IfInOctets
	if err := client.Walk(oids.IfInOctets, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).InOctets = convert.PDUToUint64(pdu)
		return nil
	}); err != nil {
		return nil, err
	}

	//IfOutOctets
	if err := client.Walk(oids.IfOutOctets, func(pdu gosnmp.SnmpPDU) error {
		idx, err := convert.ParseLastIntIndex(pdu.Name)
		if err != nil {
			return err
		}
		getRow(idx).OutOctets = convert.PDUToUint64(pdu)
		return nil
	}); err != nil {
		return nil, err
	}

	out := make([]IfRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out, nil
}

/* ---------------- IP Routes ---------------- */

type RouteRow struct {
	Dest    string
	Mask    string
	NextHop string
	IfIndex int
	Type    int
}

func collectIPRoutes(client *snmp.Client, oids config.IPRoutesOIDs) ([]RouteRow, error) {
	rows := map[string]*RouteRow{}
	getRow := func(dest string) *RouteRow {
		r, ok := rows[dest]
		if !ok {
			r = &RouteRow{Dest: dest}
			rows[dest] = r
		}
		return r
	}

	if err := client.Walk(oids.IpRouteDest, func(pdu gosnmp.SnmpPDU) error {
		dest, err := convert.ParseLastIPv4FromOID(pdu.Name)
		if err != nil {
			return err
		}
		getRow(dest)
		return nil
	}); err != nil {
		return nil, err
	}

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
	return out, nil
}
