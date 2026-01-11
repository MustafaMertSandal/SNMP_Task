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

\restrict A9iIdwACkWawm0mnbgv0VF02jKxIPQo7Hqt62JL8ZbExrNIgTNIpixpLS7iweRe

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

\unrestrict A9iIdwACkWawm0mnbgv0VF02jKxIPQo7Hqt62JL8ZbExrNIgTNIpixpLS7iweRe

