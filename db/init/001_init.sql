/*
   0) TimescaleDB Extension
*/
CREATE EXTENSION IF NOT EXISTS timescaledb;

ALTER DATABASE tsdb SET TIME ZONE 'Europe/Istanbul';
/*
   1) Devices
*/
CREATE TABLE IF NOT EXISTS devices (
  device_id   BIGSERIAL PRIMARY KEY,
  device_name TEXT NOT NULL UNIQUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

/*
   2) router_snmp (hypertable)
 */
CREATE TABLE IF NOT EXISTS router_snmp (
  time           TIMESTAMPTZ NOT NULL,
  device_id      BIGINT NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,

  sys_name       TEXT NOT NULL,
  sys_uptime_cs  BIGINT NOT NULL,   -- TimeTicks => 1/100 saniye
  sys_descr      TEXT NOT NULL,

  PRIMARY KEY (device_id, time)
);

-- Chunk interval: 1 minute (retention 10 dk için en pratik yaklaşım)
SELECT create_hypertable(
  'router_snmp', 'time',
  if_not_exists => TRUE,
  chunk_time_interval => INTERVAL '1 minute'
);

CREATE INDEX IF NOT EXISTS idx_router_snmp_device_time
  ON router_snmp (device_id, time DESC);

/*
   3) router_interface_metrics (hypertable)
*/
CREATE TABLE IF NOT EXISTS router_interface_metrics (
  time           TIMESTAMPTZ NOT NULL,
  device_id      BIGINT NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,

  if_index       INT NOT NULL,
  if_descr       TEXT NOT NULL,
  if_oper_status SMALLINT NOT NULL, -- 1 up, 2 down, 3 testing...
  if_in_octets   BIGINT NOT NULL,
  if_out_octets  BIGINT NOT NULL,

  PRIMARY KEY (device_id, if_index, time)
);

SELECT create_hypertable(
  'router_interface_metrics', 'time',
  if_not_exists => TRUE,
  chunk_time_interval => INTERVAL '1 minute'
);

CREATE INDEX IF NOT EXISTS idx_if_metrics_device_if_time
  ON router_interface_metrics (device_id, if_index, time DESC);

/*
   4) router_ip_routes (hypertable)
*/
CREATE TABLE IF NOT EXISTS router_ip_routes (
  time       TIMESTAMPTZ NOT NULL,
  device_id  BIGINT NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,

  dest       INET NOT NULL,
  mask       INET NOT NULL,
  next_hop   INET,
  if_index   INT,
  route_type SMALLINT,

  PRIMARY KEY (device_id, time, dest, mask, next_hop, if_index, route_type)
);

SELECT create_hypertable(
  'router_ip_routes', 'time',
  if_not_exists => TRUE,
  chunk_time_interval => INTERVAL '1 minute'
);

CREATE INDEX IF NOT EXISTS idx_routes_device_time
  ON router_ip_routes (device_id, time DESC);


/*
   5) Retention Policy (10 dk)
   - add_retention_policy => drop_chunks (çok hızlı)
   - schedule_interval'ı 1 dk yaptım (10 dk hedefi için)
*/
SELECT add_retention_policy(
  'router_snmp',
  INTERVAL '10 minutes',
  schedule_interval => INTERVAL '1 minute',
  if_not_exists => TRUE
);

SELECT add_retention_policy(
  'router_interface_metrics',
  INTERVAL '10 minutes',
  schedule_interval => INTERVAL '1 minute',
  if_not_exists => TRUE
);

SELECT add_retention_policy(
  'router_ip_routes',
  INTERVAL '10 minutes',
  schedule_interval => INTERVAL '1 minute',
  if_not_exists => TRUE
);

/* 
   6) Downsample (1 dk) - Continuous Aggregates
   Not: "last(value, time)" TimescaleDB aggregate'ıdır.
*/

-- 6.1 router_snmp_1m
CREATE MATERIALIZED VIEW IF NOT EXISTS router_snmp_1m
WITH (timescaledb.continuous) AS
SELECT
  time_bucket(INTERVAL '1 minute', time) AS bucket,
  device_id,
  last(sys_name, time)      AS sys_name,
  last(sys_uptime_cs, time) AS sys_uptime_cs,
  last(sys_descr, time)     AS sys_descr
FROM router_snmp
GROUP BY bucket, device_id;

CREATE INDEX IF NOT EXISTS idx_router_snmp_1m_device_bucket
  ON router_snmp_1m (device_id, bucket DESC);

SELECT add_continuous_aggregate_policy(
  'router_snmp_1m',
  start_offset => INTERVAL '10 minutes',
  end_offset => INTERVAL '1 minute',
  schedule_interval => INTERVAL '1 minute',
  if_not_exists => TRUE
);

-- 6.2 router_interface_metrics_1m
CREATE MATERIALIZED VIEW IF NOT EXISTS router_interface_metrics_1m
WITH (timescaledb.continuous) AS
SELECT
  time_bucket(INTERVAL '1 minute', time) AS bucket,
  device_id,
  if_index,
  last(if_descr, time)       AS if_descr,
  last(if_oper_status, time) AS if_oper_status,
  last(if_in_octets, time)   AS if_in_octets,
  last(if_out_octets, time)  AS if_out_octets
FROM router_interface_metrics
GROUP BY bucket, device_id, if_index;

CREATE INDEX IF NOT EXISTS idx_if_metrics_1m_device_if_bucket
  ON router_interface_metrics_1m (device_id, if_index, bucket DESC);

SELECT add_continuous_aggregate_policy(
  'router_interface_metrics_1m',
  start_offset => INTERVAL '10 minutes',
  end_offset => INTERVAL '1 minute',
  schedule_interval => INTERVAL '1 minute',
  if_not_exists => TRUE
);

-- 6.3 router_ip_routes_1m
-- (Aynı route key’i için dakikadaki son snapshot’ı saklar)
CREATE MATERIALIZED VIEW IF NOT EXISTS router_ip_routes_1m
WITH (timescaledb.continuous) AS
SELECT
  time_bucket(INTERVAL '1 minute', time) AS bucket,
  device_id,
  dest,
  mask,
  next_hop,
  if_index,
  route_type,
  last(time, time) AS last_seen_time
FROM router_ip_routes
GROUP BY bucket, device_id, dest, mask, next_hop, if_index, route_type;

CREATE INDEX IF NOT EXISTS idx_routes_1m_device_bucket
  ON router_ip_routes_1m (device_id, bucket DESC);

SELECT add_continuous_aggregate_policy(
  'router_ip_routes_1m',
  start_offset => INTERVAL '10 minutes',
  end_offset => INTERVAL '1 minute',
  schedule_interval => INTERVAL '1 minute',
  if_not_exists => TRUE
);

/*
   7) Retention Policy (20 dk)
   - add_retention_policy 
   - schedule_interval'ı 1 dk yaptım (10 dk hedefi için)
*/

SELECT add_retention_policy(
	'router_snmp_1m', 
	INTERVAL '20 minutes',
	schedule_interval => INTERVAL '1 minute', 
	if_not_exists => TRUE
);

SELECT add_retention_policy(
	'router_interface_metrics_1m', 
	INTERVAL '20 minutes',
	schedule_interval => INTERVAL '1 minute', 
	if_not_exists => TRUE
);

SELECT add_retention_policy(
	'router_ip_routes_1m', 
	INTERVAL '20 minutes',
	schedule_interval => INTERVAL '1 minute', 
	if_not_exists => TRUE
);