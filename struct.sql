--
-- PostgreSQL database dump
--

-- Dumped from database version 9.3.2
-- Dumped by pg_dump version 9.5.1

-- Started on 2017-07-03 11:30:24 CST

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

DROP DATABASE pyramid;
--
-- TOC entry 2258 (class 1262 OID 175635)
-- Name: pyramid; Type: DATABASE; Schema: -; Owner: -
--

CREATE DATABASE pyramid WITH TEMPLATE = template0 ENCODING = 'UTF8' LC_COLLATE = 'zh_CN.UTF-8' LC_CTYPE = 'zh_CN.UTF-8';


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
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- TOC entry 2259 (class 0 OID 0)
-- Dependencies: 8
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- TOC entry 1 (class 3079 OID 12018)
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- TOC entry 2261 (class 0 OID 0)
-- Dependencies: 1
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- TOC entry 2 (class 3079 OID 175636)
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- TOC entry 2262 (class 0 OID 0)
-- Dependencies: 2
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- TOC entry 172 (class 1259 OID 175647)
-- Name: accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE accounts (
    id uuid NOT NULL,
    member_id uuid NOT NULL,
    amount numeric(11,2) DEFAULT (0)::numeric NOT NULL,
    expiredate date
);


--
-- TOC entry 176 (class 1259 OID 175663)
-- Name: members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE members (
    id uuid NOT NULL,
    cardno text,
    phone text,
    level integer,
    createtime timestamp without time zone,
    reference_id uuid
);


--
-- TOC entry 2263 (class 0 OID 0)
-- Dependencies: 176
-- Name: COLUMN members.cardno; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN members.cardno IS '卡号';


--
-- TOC entry 2264 (class 0 OID 0)
-- Dependencies: 176
-- Name: COLUMN members.phone; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN members.phone IS '手机';


--
-- TOC entry 2265 (class 0 OID 0)
-- Dependencies: 176
-- Name: COLUMN members.level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN members.level IS '用户等级';


--
-- TOC entry 173 (class 1259 OID 175651)
-- Name: systemsettings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE systemsettings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 174 (class 1259 OID 175653)
-- Name: system_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE system_settings (
    id integer DEFAULT nextval('systemsettings_id_seq'::regclass) NOT NULL,
    code text NOT NULL,
    value text NOT NULL,
    remark text,
    updtime timestamp without time zone NOT NULL
);


--
-- TOC entry 175 (class 1259 OID 175660)
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE transactions (
    id uuid NOT NULL,
    order_id uuid,
    source_id uuid NOT NULL,
    target_id uuid NOT NULL,
    amount numeric(11,2) NOT NULL,
    transactiontime timestamp without time zone NOT NULL
);


--
-- TOC entry 2266 (class 0 OID 0)
-- Dependencies: 175
-- Name: TABLE transactions; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE transactions IS '交易流水,获取金额或消费金额';


--
-- TOC entry 177 (class 1259 OID 175669)
-- Name: user_level_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE user_level_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 178 (class 1259 OID 175671)
-- Name: user_level; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE user_level (
    id integer DEFAULT nextval('user_level_id_seq'::regclass) NOT NULL,
    sonnode_id uuid NOT NULL,
    ancestornode_id uuid,
    royaltyratio numeric(5,4) DEFAULT 0 NOT NULL,
    generations integer NOT NULL
);


--
-- TOC entry 2247 (class 0 OID 175647)
-- Dependencies: 172
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- TOC entry 2251 (class 0 OID 175663)
-- Dependencies: 176
-- Data for Name: members; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- TOC entry 2249 (class 0 OID 175653)
-- Dependencies: 174
-- Data for Name: system_settings; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (1, 'levels', '3', '提成层数', '2017-06-06 09:50:27.641084');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (2, 'level0ratio', '0.1', '第0层分成比例', '2017-06-06 09:52:01');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (3, 'level1ratio', '0.05', '第1层分成比例', '2017-06-06 09:51:00');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (4, 'level2ratio', '0.03', '第2层分成比例', '2017-06-06 09:51:01');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (5, 'level3ratio', '0.02', '第3层分成比例', '2017-06-06 09:52:01');
INSERT INTO system_settings (id, code, value, remark, updtime) VALUES (6, 'NewUsereBonus', '500', '单位人民币分', '2017-06-06 09:52:01');


--
-- TOC entry 2267 (class 0 OID 0)
-- Dependencies: 173
-- Name: systemsettings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('systemsettings_id_seq', 4, true);


--
-- TOC entry 2250 (class 0 OID 175660)
-- Dependencies: 175
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- TOC entry 2253 (class 0 OID 175671)
-- Dependencies: 178
-- Data for Name: user_level; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- TOC entry 2268 (class 0 OID 0)
-- Dependencies: 177
-- Name: user_level_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('user_level_id_seq', 1, false);


--
-- TOC entry 2125 (class 2606 OID 175713)
-- Name: accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);


--
-- TOC entry 2131 (class 2606 OID 175677)
-- Name: members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY members
    ADD CONSTRAINT members_pkey PRIMARY KEY (id);


--
-- TOC entry 2127 (class 2606 OID 175715)
-- Name: system_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY system_settings
    ADD CONSTRAINT system_settings_pkey PRIMARY KEY (id);


--
-- TOC entry 2129 (class 2606 OID 175709)
-- Name: transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- TOC entry 2133 (class 2606 OID 175711)
-- Name: user_level_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY user_level
    ADD CONSTRAINT user_level_pkey PRIMARY KEY (id);


--
-- TOC entry 2134 (class 2606 OID 175693)
-- Name: accounts_member_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY accounts
    ADD CONSTRAINT accounts_member_id_fkey FOREIGN KEY (member_id) REFERENCES members(id);


--
-- TOC entry 2137 (class 2606 OID 175678)
-- Name: fkmember_reference; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY members
    ADD CONSTRAINT member_reference_id_fkey FOREIGN KEY (reference_id) REFERENCES members(id);


--
-- TOC entry 2135 (class 2606 OID 175698)
-- Name: transactions_source_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY transactions
    ADD CONSTRAINT transactions_source_id_fkey FOREIGN KEY (source_id) REFERENCES members(id);


--
-- TOC entry 2136 (class 2606 OID 175703)
-- Name: transactions_target_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY transactions
    ADD CONSTRAINT transactions_target_id_fkey FOREIGN KEY (target_id) REFERENCES members(id);


--
-- TOC entry 2139 (class 2606 OID 175688)
-- Name: user_level_ancestornode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY user_level
    ADD CONSTRAINT user_level_ancestornode_id_fkey FOREIGN KEY (ancestornode_id) REFERENCES members(id);


--
-- TOC entry 2138 (class 2606 OID 175683)
-- Name: user_level_sonnode_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY user_level
    ADD CONSTRAINT user_level_sonnode_id_fkey FOREIGN KEY (sonnode_id) REFERENCES members(id);


--
-- TOC entry 2260 (class 0 OID 0)
-- Dependencies: 8
-- Name: public; Type: ACL; Schema: -; Owner: -
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM ydy;
GRANT ALL ON SCHEMA public TO ydy;
GRANT ALL ON SCHEMA public TO PUBLIC;


-- Completed on 2017-07-03 11:30:24 CST

--
-- PostgreSQL database dump complete
--

