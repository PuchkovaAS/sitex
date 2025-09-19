--
-- PostgreSQL database dump
--

-- Dumped from database version 16.4 (Debian 16.4-1.pgdg120+2)
-- Dumped by pg_dump version 16.4 (Debian 16.4-1.pgdg120+2)

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

ALTER TABLE IF EXISTS ONLY public.status_periods DROP CONSTRAINT IF EXISTS fk_status_types_periods;
ALTER TABLE IF EXISTS ONLY public.status_periods DROP CONSTRAINT IF EXISTS fk_status_periods_who_added;
ALTER TABLE IF EXISTS ONLY public.status_periods DROP CONSTRAINT IF EXISTS fk_status_periods_status_type;
ALTER TABLE IF EXISTS ONLY public.status_periods DROP CONSTRAINT IF EXISTS fk_status_periods_status;
ALTER TABLE IF EXISTS ONLY public.status_periods DROP CONSTRAINT IF EXISTS fk_status_periods_employee;
ALTER TABLE IF EXISTS ONLY public.status_periods DROP CONSTRAINT IF EXISTS fk_employees_status_periods;
ALTER TABLE IF EXISTS ONLY public.employees DROP CONSTRAINT IF EXISTS fk_departments_employees;
DROP INDEX IF EXISTS public.idx_status_types_deleted_at;
DROP INDEX IF EXISTS public.idx_status_periods_status;
DROP INDEX IF EXISTS public.idx_status_periods_employee_date;
DROP INDEX IF EXISTS public.idx_status_periods_deleted_at;
DROP INDEX IF EXISTS public.idx_employees_email;
DROP INDEX IF EXISTS public.idx_employees_department_id;
DROP INDEX IF EXISTS public.idx_employees_deleted_at;
DROP INDEX IF EXISTS public.idx_employees_admin;
DROP INDEX IF EXISTS public.idx_employees_active;
DROP INDEX IF EXISTS public.idx_departments_deleted_at;
DROP INDEX IF EXISTS public.idx_department_name;
DROP INDEX IF EXISTS public.e;
ALTER TABLE IF EXISTS ONLY public.departments DROP CONSTRAINT IF EXISTS uni_departments_name;
ALTER TABLE IF EXISTS ONLY public.status_types DROP CONSTRAINT IF EXISTS status_types_pkey;
ALTER TABLE IF EXISTS ONLY public.status_periods DROP CONSTRAINT IF EXISTS status_periods_pkey;
ALTER TABLE IF EXISTS ONLY public.sessions DROP CONSTRAINT IF EXISTS sessions_pkey;
ALTER TABLE IF EXISTS ONLY public.employees DROP CONSTRAINT IF EXISTS employees_pkey;
ALTER TABLE IF EXISTS ONLY public.departments DROP CONSTRAINT IF EXISTS departments_pkey;
ALTER TABLE IF EXISTS public.status_types ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS public.status_periods ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS public.employees ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS public.departments ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.status_types_id_seq;
DROP TABLE IF EXISTS public.status_types;
DROP SEQUENCE IF EXISTS public.status_periods_id_seq;
DROP TABLE IF EXISTS public.status_periods;
DROP TABLE IF EXISTS public.sessions;
DROP SEQUENCE IF EXISTS public.employees_id_seq;
DROP TABLE IF EXISTS public.employees;
DROP SEQUENCE IF EXISTS public.departments_id_seq;
DROP TABLE IF EXISTS public.departments;
SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: departments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.departments (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL
);


ALTER TABLE public.departments OWNER TO postgres;

--
-- Name: departments_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.departments_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.departments_id_seq OWNER TO postgres;

--
-- Name: departments_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.departments_id_seq OWNED BY public.departments.id;


--
-- Name: employees; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.employees (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    role text DEFAULT 'employee'::text NOT NULL,
    "position" text,
    is_active boolean DEFAULT true,
    is_admin boolean DEFAULT false,
    department_id bigint
);


ALTER TABLE public.employees OWNER TO postgres;

--
-- Name: employees_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.employees_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.employees_id_seq OWNER TO postgres;

--
-- Name: employees_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.employees_id_seq OWNED BY public.employees.id;


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sessions (
    k character varying(64) DEFAULT ''::character varying NOT NULL,
    v bytea NOT NULL,
    e bigint DEFAULT '0'::bigint NOT NULL
);


ALTER TABLE public.sessions OWNER TO postgres;

--
-- Name: status_periods; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.status_periods (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    employee_id bigint NOT NULL,
    status_id bigint NOT NULL,
    start_date timestamp with time zone NOT NULL,
    comment text,
    one_time_event boolean DEFAULT false NOT NULL,
    who_added_id bigint NOT NULL
);


ALTER TABLE public.status_periods OWNER TO postgres;

--
-- Name: status_periods_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.status_periods_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.status_periods_id_seq OWNER TO postgres;

--
-- Name: status_periods_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.status_periods_id_seq OWNED BY public.status_periods.id;


--
-- Name: status_types; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.status_types (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    code text NOT NULL,
    one_time_event boolean DEFAULT false NOT NULL
);


ALTER TABLE public.status_types OWNER TO postgres;

--
-- Name: status_types_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.status_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.status_types_id_seq OWNER TO postgres;

--
-- Name: status_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.status_types_id_seq OWNED BY public.status_types.id;


--
-- Name: departments id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.departments ALTER COLUMN id SET DEFAULT nextval('public.departments_id_seq'::regclass);


--
-- Name: employees id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees ALTER COLUMN id SET DEFAULT nextval('public.employees_id_seq'::regclass);


--
-- Name: status_periods id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods ALTER COLUMN id SET DEFAULT nextval('public.status_periods_id_seq'::regclass);


--
-- Name: status_types id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_types ALTER COLUMN id SET DEFAULT nextval('public.status_types_id_seq'::regclass);


--
-- Data for Name: departments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.departments (id, created_at, updated_at, deleted_at, name) FROM stdin;
1	2025-09-07 04:20:20.388054+00	2025-09-07 04:20:20.388054+00	\N	УКП
2	2025-09-07 04:44:46.991469+00	2025-09-07 04:44:46.991469+00	\N	AA
3	2025-09-07 04:50:07.383666+00	2025-09-07 04:50:07.383666+00	\N	1
\.


--
-- Data for Name: employees; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.employees (id, created_at, updated_at, deleted_at, first_name, last_name, email, password_hash, role, "position", is_active, is_admin, department_id) FROM stdin;
2	2025-09-03 14:36:49.061791+00	2025-09-12 13:12:10.412721+00	\N	Александр	Кислицин	kaa@ya.ru	$2a$10$efsVqXghMXQvZe3kpuXWbewzb7qmOAkdTAOVeAR50vsGpAvXywV4C	employee	Заместитель начальника отдела	t	t	1
1	2025-08-21 10:59:42.280463+00	2025-09-19 13:23:10.134277+00	\N	Анастасия	Пучкова	a@a.ru	$2a$10$ibKRQ.daiCT1q8SQylmlGuTXyOhKa.189e1dQ5FjAtd6c8cPVXsTO	admin	Ведущий инженер	t	t	1
3	2025-09-07 04:50:07.386884+00	2025-09-07 04:50:07.386884+00	\N	Андрей	Сорокин	slav@ya.ru	$2a$10$ttqgo4whFSROHYs.laN9ues6lGZBNwASZXm3ALjJf2eXM1ybM5Mq6	employee	Начальник отдела	t	t	3
4	2025-09-07 11:46:42.12839+00	2025-09-07 11:47:13.257664+00	\N	Александр	Левашов	lev@ya.ru	$2a$10$l1fHgvLKISyUacF5O4FhPu0miXwrlGxzZpaVDvukiJHy75uZAfoIO	employee	Глав спец	t	f	1
5	2025-09-07 11:54:23.109166+00	2025-09-09 14:10:03.682861+00	\N	Иван	Иванов	a1@a.ru	$2a$10$PxUZOI4AnFpcfUjoSTRrWOLjkI/EXfsxaHph1R8u0Oxr7Qh9xf2YW	employee	asd	f	t	1
\.


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sessions (k, v, e) FROM stdin;
7e9bc5f8-ca83-46bc-9a26-015f8b461805	\\x0d7f040102ff8000010c0110000004ff800000	1758370277
68052ce0-9381-49f3-90ac-4455f8cb27d5	\\x0d7f040102ff8000010c011000002dff80000205656d61696c06737472696e670c0800066140612e72750869735f61646d696e04626f6f6c02020001	1758370511
121276f9-c23a-4c12-af1e-2bac1a8d29b7	\\x0d7f040102ff8000010c0110000004ff800000	1758370892
80d3921d-cc82-46ba-a90c-d97b77ca5683	\\x0d7f040102ff8000010c0110000016ff8000010869735f61646d696e04626f6f6c02020001	1758374597
4b58b845-d1ca-473f-b2af-7bae21d82d9f	\\x0d7f040102ff8000010c011000001bff80000105656d61696c06737472696e670c0800066140612e7275	1758379086
\.


--
-- Data for Name: status_periods; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.status_periods (id, created_at, updated_at, deleted_at, employee_id, status_id, start_date, comment, one_time_event, who_added_id) FROM stdin;
20	2025-08-26 13:34:05.884624+00	2025-08-26 13:34:11.704224+00	\N	1	2	2025-08-13 00:00:00+00		t	1
21	2025-08-26 13:35:50.288252+00	2025-08-26 13:35:50.288252+00	\N	1	8	2025-08-16 00:00:00+00		t	1
22	2025-08-26 14:26:26.069467+00	2025-08-26 14:26:26.069467+00	\N	1	1	2025-01-01 00:00:00+00		f	1
23	2025-08-27 11:25:56.579742+00	2025-08-27 11:26:37.752352+00	\N	1	1	2025-07-28 00:00:00+00		f	1
26	2025-08-27 11:49:16.244417+00	2025-08-27 11:49:16.244417+00	\N	1	4	2025-08-10 00:00:00+00		t	1
28	2025-08-27 11:52:08.009307+00	2025-08-27 11:52:08.009307+00	\N	1	1	2025-07-27 00:00:00+00		t	1
29	2025-08-27 11:52:14.909978+00	2025-08-27 11:52:14.909978+00	\N	1	1	2025-08-01 00:00:00+00		t	1
27	2025-08-27 11:50:24.140293+00	2025-08-27 11:52:48.998205+00	\N	1	3	2025-07-01 00:00:00+00		t	1
30	2025-08-27 12:55:24.662003+00	2025-08-27 12:55:24.662003+00	\N	1	2	2025-04-01 00:00:00+00		t	1
31	2025-08-27 13:54:13.159523+00	2025-08-27 13:55:47.564558+00	\N	1	3	2025-03-01 00:00:00+00		t	1
32	2025-08-27 14:07:01.441227+00	2025-08-27 14:07:01.441227+00	\N	1	2	2025-06-03 00:00:00+00		t	1
33	2025-08-27 14:08:27.396858+00	2025-08-27 14:08:27.396858+00	\N	1	2	2025-06-07 00:00:00+00		t	1
34	2025-08-27 14:08:59.564689+00	2025-08-27 14:08:59.564689+00	\N	1	3	2025-05-04 00:00:00+00		t	1
24	2025-08-27 11:26:16.740041+00	2025-08-27 16:16:19.385369+00	\N	1	1	2025-08-27 00:00:00+00		f	1
38	2025-08-27 16:21:50.932335+00	2025-08-27 16:21:50.932335+00	\N	1	5	2025-07-02 00:00:00+00		t	1
35	2025-08-27 14:09:18.384985+00	2025-08-28 13:15:40.827657+00	\N	1	7	2025-03-04 00:00:00+00	fsd	t	1
40	2025-08-27 17:01:33.084484+00	2025-08-27 17:01:33.084484+00	2025-08-28 14:20:31.029261+00	1	1	2025-08-22 00:00:00+00		f	1
39	2025-08-27 17:01:19.517598+00	2025-08-27 17:01:19.517598+00	2025-08-28 14:20:42.175152+00	1	3	2025-08-21 00:00:00+00		f	1
36	2025-08-27 16:13:31.033746+00	2025-08-27 16:13:31.033746+00	2025-08-28 14:20:54.887112+00	1	4	2025-03-03 00:00:00+00		t	1
19	2025-08-26 13:05:24.586297+00	2025-08-27 17:00:57.482125+00	2025-08-28 14:21:03.249871+00	1	3	2025-08-26 00:00:00+00		f	1
25	2025-08-27 11:26:29.26771+00	2025-08-28 14:22:47.310054+00	\N	1	2	2025-08-28 00:00:00+00		t	1
37	2025-08-27 16:15:53.19766+00	2025-08-27 16:15:53.19766+00	2025-08-29 15:18:31.50319+00	1	3	2025-04-04 00:00:00+00		t	1
42	2025-08-29 15:20:08.662908+00	2025-08-29 15:20:08.662908+00	2025-08-30 12:02:14.240638+00	1	2	2025-08-20 00:00:00+00	ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffggggggggggggggggggggggggggggggggggggggggggggggggggggg	f	1
44	2025-09-03 12:41:50.207848+00	2025-09-03 12:41:50.207848+00	\N	1	1	2025-09-03 00:00:00+00		f	1
45	2025-09-03 14:37:44.808482+00	2025-09-03 14:37:44.808482+00	\N	2	1	2025-01-04 00:00:00+00		f	2
46	2025-09-07 04:50:40.66257+00	2025-09-07 04:50:40.66257+00	\N	3	1	2025-01-31 00:00:00+00		f	3
41	2025-08-29 15:19:23.236501+00	2025-08-29 15:19:23.236501+00	2025-09-07 12:19:59.559427+00	1	4	2025-08-30 00:00:00+00	hhhhhhhhhhhhhhhhhhhhhh	t	1
48	2025-09-09 13:46:33.19558+00	2025-09-09 13:48:00.861087+00	\N	4	2	2025-01-09 00:00:00+00		f	1
49	2025-09-09 14:08:02.391183+00	2025-09-09 14:08:02.391183+00	\N	5	1	2025-01-01 00:00:00+00		f	1
50	2025-09-09 14:10:59.882044+00	2025-09-09 14:10:59.882044+00	\N	2	3	2025-09-09 00:00:00+00		t	1
51	2025-09-10 15:08:22.383658+00	2025-09-10 15:08:22.383658+00	2025-09-10 15:15:55.877331+00	1	1	2025-09-10 00:00:00+00		t	1
52	2025-09-11 14:10:14.479002+00	2025-09-11 14:10:14.479002+00	2025-09-11 14:14:29.275968+00	4	1	2025-09-11 00:00:00+00		f	1
55	2025-09-12 12:58:17.096117+00	2025-09-12 12:58:17.096117+00	2025-09-12 13:10:21.356187+00	1	4	2025-09-11 00:00:00+00		t	1
54	2025-09-12 12:21:46.063951+00	2025-09-12 12:21:46.063951+00	2025-09-12 13:10:43.247696+00	1	2	2025-09-12 00:00:00+00		t	1
58	2025-09-12 13:12:24.455854+00	2025-09-12 13:12:24.455854+00	2025-09-12 13:13:02.867535+00	1	2	2025-09-11 00:00:00+00		t	1
57	2025-09-12 13:11:22.403484+00	2025-09-12 13:11:52.22211+00	2025-09-12 13:13:07.142047+00	1	4	2025-09-12 00:00:00+00		f	1
47	2025-09-09 13:20:33.74636+00	2025-09-09 13:32:31.40206+00	2025-09-12 13:13:10.97138+00	1	1	2025-01-09 00:00:00+00		f	1
43	2025-09-02 14:50:48.788376+00	2025-09-02 15:04:57.754216+00	2025-09-12 13:13:22.70868+00	1	5	2025-09-02 00:00:00+00		f	1
53	2025-09-12 12:20:41.512321+00	2025-09-12 12:20:41.512321+00	2025-09-12 13:13:47.358699+00	3	2	2025-09-12 00:00:00+00		t	1
56	2025-09-12 13:00:36.266272+00	2025-09-12 13:00:36.266272+00	2025-09-12 13:13:53.212641+00	3	5	2025-09-11 00:00:00+00		t	1
59	2025-09-12 13:37:57.710927+00	2025-09-12 13:37:57.710927+00	2025-09-12 13:38:10.578064+00	1	2	2025-09-12 00:00:00+00		t	1
60	2025-09-13 10:45:35.960835+00	2025-09-13 10:45:35.960835+00	\N	1	7	2025-09-13 00:00:00+00		t	1
\.


--
-- Data for Name: status_types; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.status_types (id, created_at, updated_at, deleted_at, name, code, one_time_event) FROM stdin;
1	2025-08-21 10:44:29.406799+00	2025-08-21 10:44:29.406799+00	\N	В офисе	work_office	f
2	2025-08-21 10:44:29.406799+00	2025-08-21 10:44:29.406799+00	\N	Удаленная работа	work_remote	f
3	2025-08-21 10:44:29.406799+00	2025-08-21 10:44:29.406799+00	\N	Командировка	business_trip	f
4	2025-08-21 10:44:29.406799+00	2025-08-21 10:44:29.406799+00	\N	Отпуск	vacation	f
5	2025-08-21 10:44:29.406799+00	2025-08-21 10:44:29.406799+00	\N	Больничный	sick_leave	f
6	2025-08-21 10:44:29.406799+00	2025-08-21 10:44:29.406799+00	\N	Выходной	weekend	f
7	2025-08-21 10:44:29.406799+00	2025-08-21 10:44:29.406799+00	\N	Отгул	day_off	f
8	2025-08-21 10:44:29.406+00	2025-08-21 10:44:29.406+00	\N	Работа в выходной день	weekend_work	f
\.


--
-- Name: departments_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.departments_id_seq', 3, true);


--
-- Name: employees_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.employees_id_seq', 5, true);


--
-- Name: status_periods_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.status_periods_id_seq', 60, true);


--
-- Name: status_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.status_types_id_seq', 7, true);


--
-- Name: departments departments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT departments_pkey PRIMARY KEY (id);


--
-- Name: employees employees_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT employees_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (k);


--
-- Name: status_periods status_periods_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods
    ADD CONSTRAINT status_periods_pkey PRIMARY KEY (id);


--
-- Name: status_types status_types_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_types
    ADD CONSTRAINT status_types_pkey PRIMARY KEY (id);


--
-- Name: departments uni_departments_name; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.departments
    ADD CONSTRAINT uni_departments_name UNIQUE (name);


--
-- Name: e; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX e ON public.sessions USING btree (e);


--
-- Name: idx_department_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_department_name ON public.departments USING btree (name);


--
-- Name: idx_departments_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_departments_deleted_at ON public.departments USING btree (deleted_at);


--
-- Name: idx_employees_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_employees_active ON public.employees USING btree (is_active);


--
-- Name: idx_employees_admin; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_employees_admin ON public.employees USING btree (is_admin);


--
-- Name: idx_employees_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_employees_deleted_at ON public.employees USING btree (deleted_at);


--
-- Name: idx_employees_department_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_employees_department_id ON public.employees USING btree (department_id);


--
-- Name: idx_employees_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_employees_email ON public.employees USING btree (email);


--
-- Name: idx_status_periods_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_status_periods_deleted_at ON public.status_periods USING btree (deleted_at);


--
-- Name: idx_status_periods_employee_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_status_periods_employee_date ON public.status_periods USING btree (employee_id, start_date);


--
-- Name: idx_status_periods_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_status_periods_status ON public.status_periods USING btree (status_id);


--
-- Name: idx_status_types_deleted_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_status_types_deleted_at ON public.status_types USING btree (deleted_at);


--
-- Name: employees fk_departments_employees; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.employees
    ADD CONSTRAINT fk_departments_employees FOREIGN KEY (department_id) REFERENCES public.departments(id);


--
-- Name: status_periods fk_employees_status_periods; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods
    ADD CONSTRAINT fk_employees_status_periods FOREIGN KEY (employee_id) REFERENCES public.employees(id);


--
-- Name: status_periods fk_status_periods_employee; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods
    ADD CONSTRAINT fk_status_periods_employee FOREIGN KEY (employee_id) REFERENCES public.employees(id) ON DELETE CASCADE;


--
-- Name: status_periods fk_status_periods_status; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods
    ADD CONSTRAINT fk_status_periods_status FOREIGN KEY (status_id) REFERENCES public.status_types(id) ON DELETE CASCADE;


--
-- Name: status_periods fk_status_periods_status_type; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods
    ADD CONSTRAINT fk_status_periods_status_type FOREIGN KEY (status_id) REFERENCES public.status_types(id);


--
-- Name: status_periods fk_status_periods_who_added; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods
    ADD CONSTRAINT fk_status_periods_who_added FOREIGN KEY (who_added_id) REFERENCES public.employees(id);


--
-- Name: status_periods fk_status_types_periods; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.status_periods
    ADD CONSTRAINT fk_status_types_periods FOREIGN KEY (status_id) REFERENCES public.status_types(id);


--
-- PostgreSQL database dump complete
--

