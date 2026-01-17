package db

import (
	"context"
	"time"
)

// ---- Modeller ----

type Device struct {
	DeviceID   int64
	DeviceName string
	CreatedAt  time.Time
}

type RouterSNMPRow struct {
	Time        time.Time
	DeviceID    int64
	SysName     string
	SysUptimeCS int64
	SysDescr    string
}

type InterfaceMetricRow struct {
	Time         time.Time
	DeviceID     int64
	IfIndex      int
	IfDescr      string
	IfOperStatus int16
	IfInOctets   int64
	IfOutOctets  int64
}

type RouteRowRead struct {
	Time      time.Time
	DeviceID  int64
	Dest      string // inet -> string (ör: "10.10.10.0")
	Mask      string // inet -> string (ör: "255.255.255.0")
	NextHop   string // inet -> string (ör: "10.10.10.1")
	IfIndex   int
	RouteType int16
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

// Tüm cihazları listeler.
func (s *Store) ListDevices(ctx context.Context) ([]Device, error) {
	const q = `
	SELECT device_id, device_name, created_at
	FROM devices
	ORDER BY device_id;
	`

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Device, 0, 16)
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.DeviceID, &d.DeviceName, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// ---- router_snmp ----

// Device_id için en güncel (time desc) 1 satır döner.
func (s *Store) GetLatestRouterSNMP(ctx context.Context, deviceID int64) (RouterSNMPRow, error) {
	const q = `
	SELECT time, device_id, sys_name, sys_uptime_cs, sys_descr
	FROM router_snmp
	WHERE device_id = $1
	ORDER BY time DESC
	LIMIT 1;
	`

	var r RouterSNMPRow
	err := s.pool.QueryRow(ctx, q, deviceID).Scan(
		&r.Time, &r.DeviceID, &r.SysName, &r.SysUptimeCS, &r.SysDescr,
	)
	return r, err
}

// ---- router_interface_metrics ----

// device_id için her if_index’in en güncel satırını döner (DISTINCT ON ile).
func (s *Store) GetLatestInterfaceMetricsPerIfIndex(ctx context.Context, deviceID int64, limit int64) ([]InterfaceMetricRow, error) {
	const q = `
	SELECT time, device_id, if_index, if_descr, if_oper_status, if_in_octets, if_out_octets
	FROM router_interface_metrics
	WHERE device_id = $1
	ORDER BY time DESC
	LIMIT $2;
	`

	rows, err := s.pool.Query(ctx, q, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]InterfaceMetricRow, 0, 32)
	for rows.Next() {
		var r InterfaceMetricRow
		if err := rows.Scan(
			&r.Time, &r.DeviceID, &r.IfIndex, &r.IfDescr, &r.IfOperStatus, &r.IfInOctets, &r.IfOutOctets,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---- router_ip_routes ----

// device_id için en son snapshot zamanını bulur ve o time’daki tüm route satırlarını döner.
func (s *Store) GetLatestRoutesSnapshot(ctx context.Context, deviceID int64) (snapshotTime time.Time, routes []RouteRowRead, err error) {
	// 1) En son time
	const qMax = `
	SELECT MAX(time)
	FROM router_ip_routes
	WHERE device_id = $1;
	`

	if err = s.pool.QueryRow(ctx, qMax, deviceID).Scan(&snapshotTime); err != nil {
		return time.Time{}, nil, err
	}
	// snapshotTime NULL olabilir → pgx bunu error yapmaz; ama zero time gelebilir.
	if snapshotTime.IsZero() {
		return time.Time{}, []RouteRowRead{}, nil
	}

	// 2) O time’daki tüm routes
	routes, err = s.GetRoutesAtTime(ctx, deviceID, snapshotTime)
	return snapshotTime, routes, err
}

// GetRoutesAtTime: device_id + time için route satırlarını döner.
func (s *Store) GetRoutesAtTime(ctx context.Context, deviceID int64, t time.Time) ([]RouteRowRead, error) {
	const q = `
	SELECT time, device_id, dest::text, mask::text, next_hop::text, if_index, route_type
	FROM router_ip_routes
	WHERE device_id = $1
	  AND time = $2
	ORDER BY dest, mask, next_hop, if_index, route_type;
	`

	rows, err := s.pool.Query(ctx, q, deviceID, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RouteRowRead, 0, 64)
	for rows.Next() {
		var r RouteRowRead
		if err := rows.Scan(&r.Time, &r.DeviceID, &r.Dest, &r.Mask, &r.NextHop, &r.IfIndex, &r.RouteType); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
