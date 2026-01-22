pg_dump: warning: there are circular foreign-key constraints on this table:
pg_dump: detail: hypertable
pg_dump: hint: You might not be able to restore the dump without using --disable-triggers or temporarily dropping the constraints.
pg_dump: hint: Consider using a full dump instead of a --data-only dump to avoid this problem.
pg_dump: warning: there are circular foreign-key constraints on this table:
pg_dump: detail: chunk
pg_dump: hint: You might not be able to restore the dump without using --disable-triggers or temporarily dropping the constraints.
pg_dump: hint: Consider using a full dump instead of a --data-only dump to avoid this problem.
pg_dump: warning: there are circular foreign-key constraints on this table:
pg_dump: detail: continuous_agg
pg_dump: hint: You might not be able to restore the dump without using --disable-triggers or temporarily dropping the constraints.
pg_dump: hint: Consider using a full dump instead of a --data-only dump to avoid this problem.
--
-- PostgreSQL database dump
--

\restrict N0eVcbNWKMtteZr4gRvS1DarYmu63jOXhuz9OVa9uo1GwGTeiI9rU0QRLivwTSr

-- Dumped from database version 16.11
-- Dumped by pg_dump version 16.11

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: timescaledb; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS timescaledb WITH SCHEMA public;


--
-- Name: EXTENSION timescaledb; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION timescaledb IS 'Enables scalable inserts and complex queries for time-series data (Community Edition)';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: router_snmp; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.router_snmp (
    "time" timestamp with time zone NOT NULL,
    device_id bigint NOT NULL,
    sys_name text NOT NULL,
    sys_uptime_cs bigint NOT NULL,
    sys_descr text NOT NULL
);


ALTER TABLE public.router_snmp OWNER TO postgres;

--
-- Name: _direct_view_22; Type: VIEW; Schema: _timescaledb_internal; Owner: postgres
--

CREATE VIEW _timescaledb_internal._direct_view_22 AS
 SELECT public.time_bucket('00:01:00'::interval, "time") AS bucket,
    device_id,
    public.last(sys_name, "time") AS sys_name,
    public.last(sys_uptime_cs, "time") AS sys_uptime_cs,
    public.last(sys_descr, "time") AS sys_descr
   FROM public.router_snmp
  GROUP BY (public.time_bucket('00:01:00'::interval, "time")), device_id;


ALTER VIEW _timescaledb_internal._direct_view_22 OWNER TO postgres;

--
-- Name: router_interface_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.router_interface_metrics (
    "time" timestamp with time zone NOT NULL,
    device_id bigint NOT NULL,
    if_index integer NOT NULL,
    if_descr text NOT NULL,
    if_oper_status smallint NOT NULL,
    if_in_octets bigint NOT NULL,
    if_out_octets bigint NOT NULL
);


ALTER TABLE public.router_interface_metrics OWNER TO postgres;

--
-- Name: _direct_view_23; Type: VIEW; Schema: _timescaledb_internal; Owner: postgres
--

CREATE VIEW _timescaledb_internal._direct_view_23 AS
 SELECT public.time_bucket('00:01:00'::interval, "time") AS bucket,
    device_id,
    if_index,
    public.last(if_descr, "time") AS if_descr,
    public.last(if_oper_status, "time") AS if_oper_status,
    public.last(if_in_octets, "time") AS if_in_octets,
    public.last(if_out_octets, "time") AS if_out_octets
   FROM public.router_interface_metrics
  GROUP BY (public.time_bucket('00:01:00'::interval, "time")), device_id, if_index;


ALTER VIEW _timescaledb_internal._direct_view_23 OWNER TO postgres;

--
-- Name: router_ip_routes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.router_ip_routes (
    "time" timestamp with time zone NOT NULL,
    device_id bigint NOT NULL,
    dest inet NOT NULL,
    mask inet NOT NULL,
    next_hop inet NOT NULL,
    if_index integer NOT NULL,
    route_type smallint NOT NULL
);


ALTER TABLE public.router_ip_routes OWNER TO postgres;

--
-- Name: _direct_view_24; Type: VIEW; Schema: _timescaledb_internal; Owner: postgres
--

CREATE VIEW _timescaledb_internal._direct_view_24 AS
 SELECT public.time_bucket('00:01:00'::interval, "time") AS bucket,
    device_id,
    dest,
    mask,
    next_hop,
    if_index,
    route_type,
    public.last("time", "time") AS last_seen_time
   FROM public.router_ip_routes
  GROUP BY (public.time_bucket('00:01:00'::interval, "time")), device_id, dest, mask, next_hop, if_index, route_type;


ALTER VIEW _timescaledb_internal._direct_view_24 OWNER TO postgres;

--
-- Name: _materialized_hypertable_22; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._materialized_hypertable_22 (
    bucket timestamp with time zone NOT NULL,
    device_id bigint,
    sys_name text,
    sys_uptime_cs bigint,
    sys_descr text
);


ALTER TABLE _timescaledb_internal._materialized_hypertable_22 OWNER TO postgres;

--
-- Name: _hyper_22_2060_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._hyper_22_2060_chunk (
    CONSTRAINT constraint_2060 CHECK (((bucket >= '2026-01-22 21:20:00+03'::timestamp with time zone) AND (bucket < '2026-01-22 21:30:00+03'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_22);


ALTER TABLE _timescaledb_internal._hyper_22_2060_chunk OWNER TO postgres;

--
-- Name: _hyper_22_2081_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._hyper_22_2081_chunk (
    CONSTRAINT constraint_2081 CHECK (((bucket >= '2026-01-22 21:30:00+03'::timestamp with time zone) AND (bucket < '2026-01-22 21:40:00+03'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_22);


ALTER TABLE _timescaledb_internal._hyper_22_2081_chunk OWNER TO postgres;

--
-- Name: _materialized_hypertable_23; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._materialized_hypertable_23 (
    bucket timestamp with time zone NOT NULL,
    device_id bigint,
    if_index integer,
    if_descr text,
    if_oper_status smallint,
    if_in_octets bigint,
    if_out_octets bigint
);


ALTER TABLE _timescaledb_internal._materialized_hypertable_23 OWNER TO postgres;

--
-- Name: _hyper_23_2061_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._hyper_23_2061_chunk (
    CONSTRAINT constraint_2061 CHECK (((bucket >= '2026-01-22 21:20:00+03'::timestamp with time zone) AND (bucket < '2026-01-22 21:30:00+03'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_23);


ALTER TABLE _timescaledb_internal._hyper_23_2061_chunk OWNER TO postgres;

--
-- Name: _hyper_23_2083_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._hyper_23_2083_chunk (
    CONSTRAINT constraint_2083 CHECK (((bucket >= '2026-01-22 21:30:00+03'::timestamp with time zone) AND (bucket < '2026-01-22 21:40:00+03'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_23);


ALTER TABLE _timescaledb_internal._hyper_23_2083_chunk OWNER TO postgres;

--
-- Name: _materialized_hypertable_24; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._materialized_hypertable_24 (
    bucket timestamp with time zone NOT NULL,
    device_id bigint,
    dest inet,
    mask inet,
    next_hop inet,
    if_index integer,
    route_type smallint,
    last_seen_time timestamp with time zone
);


ALTER TABLE _timescaledb_internal._materialized_hypertable_24 OWNER TO postgres;

--
-- Name: _hyper_24_2062_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._hyper_24_2062_chunk (
    CONSTRAINT constraint_2062 CHECK (((bucket >= '2026-01-22 21:20:00+03'::timestamp with time zone) AND (bucket < '2026-01-22 21:30:00+03'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_24);


ALTER TABLE _timescaledb_internal._hyper_24_2062_chunk OWNER TO postgres;

--
-- Name: _hyper_24_2082_chunk; Type: TABLE; Schema: _timescaledb_internal; Owner: postgres
--

CREATE TABLE _timescaledb_internal._hyper_24_2082_chunk (
    CONSTRAINT constraint_2082 CHECK (((bucket >= '2026-01-22 21:30:00+03'::timestamp with time zone) AND (bucket < '2026-01-22 21:40:00+03'::timestamp with time zone)))
)
INHERITS (_timescaledb_internal._materialized_hypertable_24);


ALTER TABLE _timescaledb_internal._hyper_24_2082_chunk OWNER TO postgres;

--
-- Name: _partial_view_22; Type: VIEW; Schema: _timescaledb_internal; Owner: postgres
--

CREATE VIEW _timescaledb_internal._partial_view_22 AS
 SELECT public.time_bucket('00:01:00'::interval, "time") AS bucket,
    device_id,
    public.last(sys_name, "time") AS sys_name,
    public.last(sys_uptime_cs, "time") AS sys_uptime_cs,
    public.last(sys_descr, "time") AS sys_descr
   FROM public.router_snmp
  GROUP BY (public.time_bucket('00:01:00'::interval, "time")), device_id;


ALTER VIEW _timescaledb_internal._partial_view_22 OWNER TO postgres;

--
-- Name: _partial_view_23; Type: VIEW; Schema: _timescaledb_internal; Owner: postgres
--

CREATE VIEW _timescaledb_internal._partial_view_23 AS
 SELECT public.time_bucket('00:01:00'::interval, "time") AS bucket,
    device_id,
    if_index,
    public.last(if_descr, "time") AS if_descr,
    public.last(if_oper_status, "time") AS if_oper_status,
    public.last(if_in_octets, "time") AS if_in_octets,
    public.last(if_out_octets, "time") AS if_out_octets
   FROM public.router_interface_metrics
  GROUP BY (public.time_bucket('00:01:00'::interval, "time")), device_id, if_index;


ALTER VIEW _timescaledb_internal._partial_view_23 OWNER TO postgres;

--
-- Name: _partial_view_24; Type: VIEW; Schema: _timescaledb_internal; Owner: postgres
--

CREATE VIEW _timescaledb_internal._partial_view_24 AS
 SELECT public.time_bucket('00:01:00'::interval, "time") AS bucket,
    device_id,
    dest,
    mask,
    next_hop,
    if_index,
    route_type,
    public.last("time", "time") AS last_seen_time
   FROM public.router_ip_routes
  GROUP BY (public.time_bucket('00:01:00'::interval, "time")), device_id, dest, mask, next_hop, if_index, route_type;


ALTER VIEW _timescaledb_internal._partial_view_24 OWNER TO postgres;

--
-- Name: devices; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.devices (
    device_id bigint NOT NULL,
    device_name text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.devices OWNER TO postgres;

--
-- Name: devices_device_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.devices_device_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.devices_device_id_seq OWNER TO postgres;

--
-- Name: devices_device_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.devices_device_id_seq OWNED BY public.devices.device_id;


--
-- Name: router_interface_metrics_1m; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.router_interface_metrics_1m AS
 SELECT bucket,
    device_id,
    if_index,
    if_descr,
    if_oper_status,
    if_in_octets,
    if_out_octets
   FROM _timescaledb_internal._materialized_hypertable_23;


ALTER VIEW public.router_interface_metrics_1m OWNER TO postgres;

--
-- Name: router_ip_routes_1m; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.router_ip_routes_1m AS
 SELECT bucket,
    device_id,
    dest,
    mask,
    next_hop,
    if_index,
    route_type,
    last_seen_time
   FROM _timescaledb_internal._materialized_hypertable_24;


ALTER VIEW public.router_ip_routes_1m OWNER TO postgres;

--
-- Name: router_snmp_1m; Type: VIEW; Schema: public; Owner: postgres
--

CREATE VIEW public.router_snmp_1m AS
 SELECT bucket,
    device_id,
    sys_name,
    sys_uptime_cs,
    sys_descr
   FROM _timescaledb_internal._materialized_hypertable_22;


ALTER VIEW public.router_snmp_1m OWNER TO postgres;

--
-- Name: devices device_id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.devices ALTER COLUMN device_id SET DEFAULT nextval('public.devices_device_id_seq'::regclass);


--
-- Name: devices devices_device_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.devices
    ADD CONSTRAINT devices_device_name_key UNIQUE (device_name);


--
-- Name: devices devices_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.devices
    ADD CONSTRAINT devices_pkey PRIMARY KEY (device_id);


--
-- Name: router_interface_metrics router_interface_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.router_interface_metrics
    ADD CONSTRAINT router_interface_metrics_pkey PRIMARY KEY (device_id, if_index, "time");


--
-- Name: router_ip_routes router_ip_routes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.router_ip_routes
    ADD CONSTRAINT router_ip_routes_pkey PRIMARY KEY (device_id, "time", dest, mask, next_hop, if_index, route_type);


--
-- Name: router_snmp router_snmp_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.router_snmp
    ADD CONSTRAINT router_snmp_pkey PRIMARY KEY (device_id, "time");


--
-- Name: _hyper_22_2060_chunk__materialized_hypertable_22_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_22_2060_chunk__materialized_hypertable_22_bucket_idx ON _timescaledb_internal._hyper_22_2060_chunk USING btree (bucket DESC);


--
-- Name: _hyper_22_2060_chunk__materialized_hypertable_22_device_id_buck; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_22_2060_chunk__materialized_hypertable_22_device_id_buck ON _timescaledb_internal._hyper_22_2060_chunk USING btree (device_id, bucket DESC);


--
-- Name: _hyper_22_2081_chunk__materialized_hypertable_22_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_22_2081_chunk__materialized_hypertable_22_bucket_idx ON _timescaledb_internal._hyper_22_2081_chunk USING btree (bucket DESC);


--
-- Name: _hyper_22_2081_chunk__materialized_hypertable_22_device_id_buck; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_22_2081_chunk__materialized_hypertable_22_device_id_buck ON _timescaledb_internal._hyper_22_2081_chunk USING btree (device_id, bucket DESC);


--
-- Name: _hyper_23_2061_chunk__materialized_hypertable_23_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2061_chunk__materialized_hypertable_23_bucket_idx ON _timescaledb_internal._hyper_23_2061_chunk USING btree (bucket DESC);


--
-- Name: _hyper_23_2061_chunk__materialized_hypertable_23_device_id_buck; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2061_chunk__materialized_hypertable_23_device_id_buck ON _timescaledb_internal._hyper_23_2061_chunk USING btree (device_id, bucket DESC);


--
-- Name: _hyper_23_2061_chunk__materialized_hypertable_23_if_index_bucke; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2061_chunk__materialized_hypertable_23_if_index_bucke ON _timescaledb_internal._hyper_23_2061_chunk USING btree (if_index, bucket DESC);


--
-- Name: _hyper_23_2061_chunk_idx_if_metrics_1m_device_if_bucket; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2061_chunk_idx_if_metrics_1m_device_if_bucket ON _timescaledb_internal._hyper_23_2061_chunk USING btree (device_id, if_index, bucket DESC);


--
-- Name: _hyper_23_2083_chunk__materialized_hypertable_23_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2083_chunk__materialized_hypertable_23_bucket_idx ON _timescaledb_internal._hyper_23_2083_chunk USING btree (bucket DESC);


--
-- Name: _hyper_23_2083_chunk__materialized_hypertable_23_device_id_buck; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2083_chunk__materialized_hypertable_23_device_id_buck ON _timescaledb_internal._hyper_23_2083_chunk USING btree (device_id, bucket DESC);


--
-- Name: _hyper_23_2083_chunk__materialized_hypertable_23_if_index_bucke; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2083_chunk__materialized_hypertable_23_if_index_bucke ON _timescaledb_internal._hyper_23_2083_chunk USING btree (if_index, bucket DESC);


--
-- Name: _hyper_23_2083_chunk_idx_if_metrics_1m_device_if_bucket; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_23_2083_chunk_idx_if_metrics_1m_device_if_bucket ON _timescaledb_internal._hyper_23_2083_chunk USING btree (device_id, if_index, bucket DESC);


--
-- Name: _hyper_24_2062_chunk__materialized_hypertable_24_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2062_chunk__materialized_hypertable_24_bucket_idx ON _timescaledb_internal._hyper_24_2062_chunk USING btree (bucket DESC);


--
-- Name: _hyper_24_2062_chunk__materialized_hypertable_24_dest_bucket_id; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2062_chunk__materialized_hypertable_24_dest_bucket_id ON _timescaledb_internal._hyper_24_2062_chunk USING btree (dest, bucket DESC);


--
-- Name: _hyper_24_2062_chunk__materialized_hypertable_24_device_id_buck; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2062_chunk__materialized_hypertable_24_device_id_buck ON _timescaledb_internal._hyper_24_2062_chunk USING btree (device_id, bucket DESC);


--
-- Name: _hyper_24_2062_chunk__materialized_hypertable_24_if_index_bucke; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2062_chunk__materialized_hypertable_24_if_index_bucke ON _timescaledb_internal._hyper_24_2062_chunk USING btree (if_index, bucket DESC);


--
-- Name: _hyper_24_2062_chunk__materialized_hypertable_24_mask_bucket_id; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2062_chunk__materialized_hypertable_24_mask_bucket_id ON _timescaledb_internal._hyper_24_2062_chunk USING btree (mask, bucket DESC);


--
-- Name: _hyper_24_2062_chunk__materialized_hypertable_24_next_hop_bucke; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2062_chunk__materialized_hypertable_24_next_hop_bucke ON _timescaledb_internal._hyper_24_2062_chunk USING btree (next_hop, bucket DESC);


--
-- Name: _hyper_24_2062_chunk__materialized_hypertable_24_route_type_buc; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2062_chunk__materialized_hypertable_24_route_type_buc ON _timescaledb_internal._hyper_24_2062_chunk USING btree (route_type, bucket DESC);


--
-- Name: _hyper_24_2082_chunk__materialized_hypertable_24_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2082_chunk__materialized_hypertable_24_bucket_idx ON _timescaledb_internal._hyper_24_2082_chunk USING btree (bucket DESC);


--
-- Name: _hyper_24_2082_chunk__materialized_hypertable_24_dest_bucket_id; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2082_chunk__materialized_hypertable_24_dest_bucket_id ON _timescaledb_internal._hyper_24_2082_chunk USING btree (dest, bucket DESC);


--
-- Name: _hyper_24_2082_chunk__materialized_hypertable_24_device_id_buck; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2082_chunk__materialized_hypertable_24_device_id_buck ON _timescaledb_internal._hyper_24_2082_chunk USING btree (device_id, bucket DESC);


--
-- Name: _hyper_24_2082_chunk__materialized_hypertable_24_if_index_bucke; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2082_chunk__materialized_hypertable_24_if_index_bucke ON _timescaledb_internal._hyper_24_2082_chunk USING btree (if_index, bucket DESC);


--
-- Name: _hyper_24_2082_chunk__materialized_hypertable_24_mask_bucket_id; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2082_chunk__materialized_hypertable_24_mask_bucket_id ON _timescaledb_internal._hyper_24_2082_chunk USING btree (mask, bucket DESC);


--
-- Name: _hyper_24_2082_chunk__materialized_hypertable_24_next_hop_bucke; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2082_chunk__materialized_hypertable_24_next_hop_bucke ON _timescaledb_internal._hyper_24_2082_chunk USING btree (next_hop, bucket DESC);


--
-- Name: _hyper_24_2082_chunk__materialized_hypertable_24_route_type_buc; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _hyper_24_2082_chunk__materialized_hypertable_24_route_type_buc ON _timescaledb_internal._hyper_24_2082_chunk USING btree (route_type, bucket DESC);


--
-- Name: _materialized_hypertable_22_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_22_bucket_idx ON _timescaledb_internal._materialized_hypertable_22 USING btree (bucket DESC);


--
-- Name: _materialized_hypertable_22_device_id_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_22_device_id_bucket_idx ON _timescaledb_internal._materialized_hypertable_22 USING btree (device_id, bucket DESC);


--
-- Name: _materialized_hypertable_23_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_23_bucket_idx ON _timescaledb_internal._materialized_hypertable_23 USING btree (bucket DESC);


--
-- Name: _materialized_hypertable_23_device_id_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_23_device_id_bucket_idx ON _timescaledb_internal._materialized_hypertable_23 USING btree (device_id, bucket DESC);


--
-- Name: _materialized_hypertable_23_if_index_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_23_if_index_bucket_idx ON _timescaledb_internal._materialized_hypertable_23 USING btree (if_index, bucket DESC);


--
-- Name: _materialized_hypertable_24_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_24_bucket_idx ON _timescaledb_internal._materialized_hypertable_24 USING btree (bucket DESC);


--
-- Name: _materialized_hypertable_24_dest_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_24_dest_bucket_idx ON _timescaledb_internal._materialized_hypertable_24 USING btree (dest, bucket DESC);


--
-- Name: _materialized_hypertable_24_device_id_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_24_device_id_bucket_idx ON _timescaledb_internal._materialized_hypertable_24 USING btree (device_id, bucket DESC);


--
-- Name: _materialized_hypertable_24_if_index_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_24_if_index_bucket_idx ON _timescaledb_internal._materialized_hypertable_24 USING btree (if_index, bucket DESC);


--
-- Name: _materialized_hypertable_24_mask_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_24_mask_bucket_idx ON _timescaledb_internal._materialized_hypertable_24 USING btree (mask, bucket DESC);


--
-- Name: _materialized_hypertable_24_next_hop_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_24_next_hop_bucket_idx ON _timescaledb_internal._materialized_hypertable_24 USING btree (next_hop, bucket DESC);


--
-- Name: _materialized_hypertable_24_route_type_bucket_idx; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX _materialized_hypertable_24_route_type_bucket_idx ON _timescaledb_internal._materialized_hypertable_24 USING btree (route_type, bucket DESC);


--
-- Name: idx_if_metrics_1m_device_if_bucket; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX idx_if_metrics_1m_device_if_bucket ON _timescaledb_internal._materialized_hypertable_23 USING btree (device_id, if_index, bucket DESC);


--
-- Name: idx_router_snmp_1m_device_bucket; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX idx_router_snmp_1m_device_bucket ON _timescaledb_internal._materialized_hypertable_22 USING btree (device_id, bucket DESC);


--
-- Name: idx_routes_1m_device_bucket; Type: INDEX; Schema: _timescaledb_internal; Owner: postgres
--

CREATE INDEX idx_routes_1m_device_bucket ON _timescaledb_internal._materialized_hypertable_24 USING btree (device_id, bucket DESC);


--
-- Name: idx_if_metrics_device_if_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_if_metrics_device_if_time ON public.router_interface_metrics USING btree (device_id, if_index, "time" DESC);


--
-- Name: idx_router_snmp_device_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_router_snmp_device_time ON public.router_snmp USING btree (device_id, "time" DESC);


--
-- Name: idx_routes_device_time; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_routes_device_time ON public.router_ip_routes USING btree (device_id, "time" DESC);


--
-- Name: router_interface_metrics_time_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX router_interface_metrics_time_idx ON public.router_interface_metrics USING btree ("time" DESC);


--
-- Name: router_ip_routes_time_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX router_ip_routes_time_idx ON public.router_ip_routes USING btree ("time" DESC);


--
-- Name: router_snmp_time_idx; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX router_snmp_time_idx ON public.router_snmp USING btree ("time" DESC);


--
-- Name: router_interface_metrics router_interface_metrics_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.router_interface_metrics
    ADD CONSTRAINT router_interface_metrics_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE;


--
-- Name: router_ip_routes router_ip_routes_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.router_ip_routes
    ADD CONSTRAINT router_ip_routes_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE;


--
-- Name: router_snmp router_snmp_device_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.router_snmp
    ADD CONSTRAINT router_snmp_device_id_fkey FOREIGN KEY (device_id) REFERENCES public.devices(device_id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict N0eVcbNWKMtteZr4gRvS1DarYmu63jOXhuz9OVa9uo1GwGTeiI9rU0QRLivwTSr

