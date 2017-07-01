--
-- PostgreSQL database dump
--

-- Dumped from database version 9.3.2
-- Dumped by pg_dump version 9.5.1

-- Started on 2017-07-01 15:25:05 CST

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

DROP DATABASE pyramid;
--
-- TOC entry 2251 (class 1262 OID 175427)
-- Name: pyramid; Type: DATABASE; Schema: -; Owner: ydy
--

CREATE DATABASE pyramid WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'zh_CN.UTF-8' LC_CTYPE = 'zh_CN.UTF-8';


ALTER DATABASE pyramid OWNER TO ydy;

\connect pyramid

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 8 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: ydy
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO ydy;

--
-- TOC entry 2252 (class 0 OID 0)
-- Dependencies: 8
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: ydy
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- TOC entry 1 (class 3079 OID 12018)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- TOC entry 2254 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 2 (class 3079 OID 175428)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 2255 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 173 (class 1259 OID 175445)
-- Name: account; Type: TABLE; Schema: public; Owner: ydy
--

CREATE TABLE account (
    id uuid NOT NULL,
    reference_id uuid NOT NULL,
    amount numeric(11,2) DEFAULT (0)::numeric NOT NULL,
    expiredate date
);


ALTER TABLE account OWNER TO ydy;

--
-- TOC entry 174 (class 1259 OID 175449)
-- Name: systemsettings_id_seq; Type: SEQUENCE; Schema: public; Owner: ydy
--

CREATE SEQUENCE systemsettings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE systemsettings_id_seq OWNER TO ydy;

--
-- TOC entry 175 (class 1259 OID 175451)
-- Name: system_settings; Type: TABLE; Schema: public; Owner: ydy
--

CREATE TABLE system_settings (
    id integer DEFAULT nextval('systemsettings_id_seq'::regclass) NOT NULL,
    code text NOT NULL,
    value text NOT NULL,
    remark text,
    updtime timestamp without time zone NOT NULL
);


ALTER TABLE system_settings OWNER TO ydy;

--
-- TOC entry 176 (class 1259 OID 175458)
-- Name: transaction; Type: TABLE; Schema: public; Owner: ydy
--

CREATE TABLE transaction (
    id uuid NOT NULL,
    order_id uuid,
    source_id uuid NOT NULL,
    target_id uuid NOT NULL,
    amount numeric(11,2) NOT NULL,
    transactiontime timestamp without time zone NOT NULL
);


ALTER TABLE transaction OWNER TO ydy;

--
-- TOC entry 2256 (class 0 OID 0)
-- Dependencies: 176
-- Name: TABLE transaction; Type: COMMENT; Schema: public; Owner: ydy
--

COMMENT ON TABLE transaction IS '交易流水,获取金额或消费金额';


--
-- TOC entry 172 (class 1259 OID 175439)
-- Name: user; Type: TABLE; Schema: public; Owner: ydy
--

CREATE TABLE member (
    id uuid NOT NULL,
    cardno text,
    phone text,
    level integer NOT NULL,
    createtime timestamp without time zone
);


ALTER TABLE member OWNER TO ydy;

--
-- TOC entry 2257 (class 0 OID 0)
-- Dependencies: 172
-- Name: COLUMN member.cardno; Type: COMMENT; Schema: public; Owner: ydy
--

COMMENT ON COLUMN member.cardno IS '卡号';


--
-- TOC entry 2258 (class 0 OID 0)
-- Dependencies: 172
-- Name: COLUMN member.phone; Type: COMMENT; Schema: public; Owner: ydy
--

COMMENT ON COLUMN member.phone IS '手机';


--
-- TOC entry 2259 (class 0 OID 0)
-- Dependencies: 172
-- Name: COLUMN member.level; Type: COMMENT; Schema: public; Owner: ydy
--

COMMENT ON COLUMN member.level IS '用户等级';


--
-- TOC entry 177 (class 1259 OID 175461)
-- Name: user_level_id_seq; Type: SEQUENCE; Schema: public; Owner: ydy
--

CREATE SEQUENCE user_level_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE user_level_id_seq OWNER TO ydy;

--
-- TOC entry 178 (class 1259 OID 175463)
-- Name: user_level; Type: TABLE; Schema: public; Owner: ydy
--

CREATE TABLE user_level (
    id integer DEFAULT nextval('user_level_id_seq'::regclass) NOT NULL,
    sonnode_id uuid NOT NULL,
    ancestornode_id uuid,
    royaltyratio numeric(5,4) DEFAULT 0 NOT NULL,
    generations integer NOT NULL
);


ALTER TABLE user_level OWNER TO ydy;

--
-- TOC entry 2241 (class 0 OID 175445)
-- Dependencies: 173
-- Data for Name: account; Type: TABLE DATA; Schema: public; Owner: ydy
--



--
-- TOC entry 2243 (class 0 OID 175451)
-- Dependencies: 175
-- Data for Name: system_settings; Type: TABLE DATA; Schema: public; Owner: ydy
--

INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (1, 'levels', '3', '提成层数', '2017-06-06 09:50:27.641084');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (2, 'level0ratio', '0.1', '第0层分成比例', '2017-06-06 09:52:01');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (3, 'level1ratio', '0.05', '第1层分成比例', '2017-06-06 09:51:00');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (4, 'level2ratio', '0.03', '第2层分成比例', '2017-06-06 09:51:01');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (5, 'level3ratio', '0.02', '第3层分成比例', '2017-06-06 09:52:01');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (6, 'NewUsereBonus', '500', '单位人民币分', '2017-06-06 09:52:01');


--
-- TOC entry 2260 (class 0 OID 0)
-- Dependencies: 174
-- Name: systemsettings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: ydy
--

SELECT pg_catalog.setval('systemsettings_id_seq', 4, true);


--
-- TOC entry 2244 (class 0 OID 175458)
-- Dependencies: 176
-- Data for Name: transaction; Type: TABLE DATA; Schema: public; Owner: ydy
--



--
-- TOC entry 2240 (class 0 OID 175439)
-- Dependencies: 172
-- Data for Name: user; Type: TABLE DATA; Schema: public; Owner: ydy
--



--
-- TOC entry 2246 (class 0 OID 175463)
-- Dependencies: 178
-- Data for Name: user_level; Type: TABLE DATA; Schema: public; Owner: ydy
--



--
-- TOC entry 2261 (class 0 OID 0)
-- Dependencies: 177
-- Name: user_level_id_seq; Type: SEQUENCE SET; Schema: public; Owner: ydy
--

SELECT pg_catalog.setval('user_level_id_seq', 1, false);


--
-- TOC entry 2253 (class 0 OID 0)
-- Dependencies: 8
-- Name: public; Type: ACL; Schema: -; Owner: ydy
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM ydy;
GRANT ALL ON SCHEMA public TO ydy;
GRANT ALL ON SCHEMA public TO PUBLIC;


-- Completed on 2017-07-01 15:25:05 CST

--
-- PostgreSQL database dump complete
--

