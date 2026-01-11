package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is a small repository layer for writing SNMP samples into TimescaleDB.
type Store struct {
	pool      *pgxpool.Pool
	batchSize int
}

func NewStore(pool *pgxpool.Pool, batchSize int) *Store {
	if batchSize <= 0 {
		batchSize = 200
	}
	return &Store{pool: pool, batchSize: batchSize}
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// EnsureDevice inserts the device (by device_name) if missing, and returns its device_id.
// devices table schema (from your db/schema.sql):
//
//	device_id bigint PK (default nextval)
//	device_name text UNIQUE NOT NULL
func (s *Store) EnsureDevice(ctx context.Context, deviceName string) (int64, error) {
	const q = `
    INSERT INTO devices (device_name)
    VALUES ($1)
    ON CONFLICT (device_name)
    DO UPDATE SET device_name = EXCLUDED.device_name
    RETURNING device_id;
    `

	var id int64
	err := s.pool.QueryRow(ctx, q, deviceName).Scan(&id)
	return id, err
}

func (s *Store) InsertRouterSNMP(ctx context.Context, t time.Time, deviceID int64, sysName string, sysUptimeCS int64, sysDescr string) error {
	const q = `
    INSERT INTO router_snmp (time, device_id, sys_name, sys_uptime_cs, sys_descr)
    VALUES ($1, $2, $3, $4, $5);
    `
	_, err := s.pool.Exec(ctx, q, t, deviceID, sysName, sysUptimeCS, sysDescr)
	return err
}

type IfMetric struct {
	IfIndex    int
	IfDescr    string
	OperStatus int
	InOctets   uint64
	OutOctets  uint64
}

// InsertInterfaceMetrics writes interface rows with pgx Batch (fast, fewer round-trips).
func (s *Store) InsertInterfaceMetrics(ctx context.Context, t time.Time, deviceID int64, rows []IfMetric) error {
	if len(rows) == 0 {
		return nil
	}

	const q = `
    INSERT INTO router_interface_metrics
    (time, device_id, if_index, if_descr, if_oper_status, if_in_octets, if_out_octets)
    VALUES ($1,$2,$3,$4,$5,$6,$7);
    `

	// Chunk into batches to avoid huge batches
	for start := 0; start < len(rows); start += s.batchSize {
		end := start + s.batchSize
		if end > len(rows) {
			end = len(rows)
		}

		b := &pgx.Batch{}
		for _, r := range rows[start:end] {
			b.Queue(q, t, deviceID, r.IfIndex, r.IfDescr, int16(r.OperStatus), int64(r.InOctets), int64(r.OutOctets))
		}

		br := s.pool.SendBatch(ctx, b)
		for range rows[start:end] {
			if _, err := br.Exec(); err != nil {
				_ = br.Close()
				return err
			}
		}
		if err := br.Close(); err != nil {
			return err
		}
	}

	return nil
}

type RouteRow struct {
	Dest    string // inet literal e.g. "10.10.10.0"
	Mask    string // inet literal e.g. "255.255.255.0"
	NextHop string // inet literal e.g. "10.10.10.1"
	IfIndex int
	Type    int
}

// InsertIPRoutes writes the current routing table snapshot as a batch.
func (s *Store) InsertIPRoutes(ctx context.Context, t time.Time, deviceID int64, routes []RouteRow) error {
	if len(routes) == 0 {
		return nil
	}

	const q = `
    INSERT INTO router_ip_routes
    (time, device_id, dest, mask, next_hop, if_index, route_type)
    VALUES ($1,$2,$3::inet,$4::inet,$5::inet,$6,$7);
    `

	for start := 0; start < len(routes); start += s.batchSize {
		end := start + s.batchSize
		if end > len(routes) {
			end = len(routes)
		}

		b := &pgx.Batch{}
		for _, r := range routes[start:end] {
			b.Queue(q, t, deviceID, r.Dest, r.Mask, r.NextHop, r.IfIndex, int16(r.Type))
		}

		br := s.pool.SendBatch(ctx, b)
		for range routes[start:end] {
			if _, err := br.Exec(); err != nil {
				_ = br.Close()
				return err
			}
		}
		if err := br.Close(); err != nil {
			return err
		}
	}

	return nil
}
