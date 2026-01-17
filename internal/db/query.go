package db

import (
	"context"
	"fmt"
	"time"
)

// ---- Modeller ----

type Device struct {
	DeviceID   int64
	DeviceName string
	CreatedAt  time.Time
}

type RouterSNMP struct {
	Bucket      time.Time
	DeviceID    int64
	SysName     string
	SysUptimeCS int64
	SysDescr    string
}

type RouterInterfaceMetric struct {
	Bucket       time.Time
	DeviceID     int64
	IfIndex      int32
	IfDescr      string
	IfOperStatus int16
	IfInOctets   int64
	IfOutOctets  int64
}

type RouterIPRoute struct {
	Bucket       time.Time
	DeviceID     int64
	Dest         string
	Mask         string
	NextHop      string
	IfIndex      int32
	RouteType    int16
	LastSeenTime time.Time
}

// ---- Devices ----

// Devices tablosundan tek cihazı getirir.
func (s *Store) GetDeviceByID(ctx context.Context, deviceID int64) (Device, error) {
	const q = `
	SELECT device_id, device_name, created_at
	FROM devices
	WHERE device_id = $1;
	`

	var d Device
	err := s.pool.QueryRow(ctx, q, deviceID).Scan(&d.DeviceID, &d.DeviceName, &d.CreatedAt)
	return d, err
}

// ---- Router_snmp ----

// Device_id için en güncel downsample edilmiş satırları döner.

func (s *Store) GetRouterSNMP1mAllByDeviceID(ctx context.Context, deviceID int64) ([]RouterSNMP, error) {
	const q = `
		SELECT
			bucket,
			device_id,
			sys_name,
			sys_uptime_cs,
			sys_descr
		FROM router_snmp_1m
		WHERE device_id = $1
		ORDER BY bucket DESC;
	`

	rows, err := s.pool.Query(ctx, q, deviceID)
	if err != nil {
		return nil, fmt.Errorf("query router_snmp_1m: %w", err)
	}
	defer rows.Close()

	out := make([]RouterSNMP, 0, 256)
	for rows.Next() {
		var r RouterSNMP
		if err := rows.Scan(
			&r.Bucket,
			&r.DeviceID,
			&r.SysName,
			&r.SysUptimeCS,
			&r.SysDescr,
		); err != nil {
			return nil, fmt.Errorf("scan router_snmp_1m: %w", err)
		}
		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows router_snmp_1m: %w", err)
	}

	return out, nil
}

// ---- Router_interface_metrics ----

// Device_id için en güncel downsample edilmiş satırları döner.
func (s *Store) GetRouterInterfaceMetrics1mAllByDeviceID(ctx context.Context, deviceID int64) ([]RouterInterfaceMetric, error) {
	const q = `
		SELECT
			bucket,
			device_id,
			if_index,
			if_descr,
			if_oper_status,
			if_in_octets,
			if_out_octets
		FROM router_interface_metrics_1m
		WHERE device_id = $1
		ORDER BY bucket DESC, if_index ASC;
	`

	rows, err := s.pool.Query(ctx, q, deviceID)
	if err != nil {
		return nil, fmt.Errorf("query router_interface_metrics_1m: %w", err)
	}
	defer rows.Close()

	out := make([]RouterInterfaceMetric, 0, 2048)

	for rows.Next() {
		var r RouterInterfaceMetric
		if err := rows.Scan(
			&r.Bucket,
			&r.DeviceID,
			&r.IfIndex,
			&r.IfDescr,
			&r.IfOperStatus,
			&r.IfInOctets,
			&r.IfOutOctets,
		); err != nil {
			return nil, fmt.Errorf("scan router_interface_metrics_1m: %w", err)
		}
		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows router_interface_metrics_1m: %w", err)
	}

	return out, nil
}

// ---- router_ip_routes ----

// Device_id için en güncel downsample edilmiş satırları döner.
func (s *Store) GetRouterIPRoutes1mAllByDeviceID(ctx context.Context, deviceID int64) ([]RouterIPRoute, error) {
	const q = `
		SELECT
			bucket,
			device_id,
			dest,
			mask,
			next_hop,
			if_index,
			route_type,
			last_seen_time
		FROM router_ip_routes_1m
		WHERE device_id = $1
		ORDER BY bucket DESC, dest, mask, next_hop, if_index;
	`

	rows, err := s.pool.Query(ctx, q, deviceID)
	if err != nil {
		return nil, fmt.Errorf("query router_ip_routes_1m: %w", err)
	}
	defer rows.Close()

	out := make([]RouterIPRoute, 0, 2048)

	for rows.Next() {
		var r RouterIPRoute
		if err := rows.Scan(
			&r.Bucket,
			&r.DeviceID,
			&r.Dest,
			&r.Mask,
			&r.NextHop,
			&r.IfIndex,
			&r.RouteType,
			&r.LastSeenTime,
		); err != nil {
			return nil, fmt.Errorf("scan router_ip_routes_1m: %w", err)
		}
		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows router_ip_routes_1m: %w", err)
	}

	return out, nil
}
