--
-- PostgreSQL database dump
--

\restrict cRrSX3qdpMuxIgMUQZSOvh0it6b0DwLHkpnBnoV6iKMcSNGwSxcFLfuSTYWIkbE

-- Dumped from database version 18.4 (Debian 18.4-1)
-- Dumped by pg_dump version 18.4 (Debian 18.4-1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: expense_category; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.expense_category AS ENUM (
    'mahsulot',
    'oylik',
    'arenda',
    'kommunal',
    'boshqa'
);


ALTER TYPE public.expense_category OWNER TO postgres;

--
-- Name: order_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.order_status AS ENUM (
    'new',
    'preparing',
    'ready',
    'on_way',
    'delivered',
    'cancelled'
);


ALTER TYPE public.order_status OWNER TO postgres;

--
-- Name: user_role; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.user_role AS ENUM (
    'customer',
    'cook',
    'courier',
    'admin'
);


ALTER TYPE public.user_role OWNER TO postgres;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: bot_staff; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.bot_staff (
    telegram_id bigint NOT NULL,
    role character varying(20) NOT NULL,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.bot_staff OWNER TO postgres;

--
-- Name: categories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.categories (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    image_url text,
    is_user_controlled boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.categories OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.categories_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.categories_id_seq OWNER TO postgres;

--
-- Name: categories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;


--
-- Name: expenses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.expenses (
    id integer NOT NULL,
    amount numeric(12,2) NOT NULL,
    category public.expense_category DEFAULT 'boshqa'::public.expense_category NOT NULL,
    description text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.expenses OWNER TO postgres;

--
-- Name: expenses_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.expenses_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.expenses_id_seq OWNER TO postgres;

--
-- Name: expenses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.expenses_id_seq OWNED BY public.expenses.id;


--
-- Name: ingredients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ingredients (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    stock numeric(12,3) DEFAULT 0 NOT NULL,
    unit character varying(20) DEFAULT 'gr'::character varying NOT NULL,
    min_stock numeric(12,3) DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.ingredients OWNER TO postgres;

--
-- Name: ingredients_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ingredients_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.ingredients_id_seq OWNER TO postgres;

--
-- Name: ingredients_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ingredients_id_seq OWNED BY public.ingredients.id;


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.order_items (
    id integer NOT NULL,
    order_id integer,
    product_id integer,
    quantity numeric(12,3) DEFAULT 1.000 NOT NULL,
    price numeric(12,2) NOT NULL,
    unit text DEFAULT 'dona'::text,
    comment text,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.order_items OWNER TO postgres;

--
-- Name: order_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.order_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.order_items_id_seq OWNER TO postgres;

--
-- Name: order_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.order_items_id_seq OWNED BY public.order_items.id;


--
-- Name: orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.orders (
    id integer NOT NULL,
    customer_id integer,
    total_price numeric(12,2) NOT NULL,
    status public.order_status DEFAULT 'new'::public.order_status,
    address text NOT NULL,
    phone character varying(20) NOT NULL,
    lat numeric(9,6),
    lng numeric(9,6),
    courier_id integer,
    cook_id integer,
    comment text,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.orders OWNER TO postgres;

--
-- Name: orders_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.orders_id_seq OWNER TO postgres;

--
-- Name: orders_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;


--
-- Name: product_ingredients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.product_ingredients (
    id integer NOT NULL,
    product_id integer,
    ingredient_id integer,
    quantity numeric(12,3) NOT NULL,
    unit character varying(20) DEFAULT 'gr'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.product_ingredients OWNER TO postgres;

--
-- Name: product_ingredients_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.product_ingredients_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.product_ingredients_id_seq OWNER TO postgres;

--
-- Name: product_ingredients_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.product_ingredients_id_seq OWNED BY public.product_ingredients.id;


--
-- Name: products; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.products (
    id integer NOT NULL,
    category_id integer,
    name character varying(255) NOT NULL,
    description text,
    price numeric(12,2) NOT NULL,
    image_url text,
    is_active boolean DEFAULT true,
    unit character varying(20) DEFAULT 'dona'::character varying,
    min_quantity numeric(12,3) DEFAULT 1.0,
    quantity_step numeric(12,3) DEFAULT 1.0,
    has_mandatory_container boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.products OWNER TO postgres;

--
-- Name: products_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.products_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.products_id_seq OWNER TO postgres;

--
-- Name: products_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;


--
-- Name: settings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.settings (
    key character varying(50) NOT NULL,
    value text NOT NULL,
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.settings OWNER TO postgres;

--
-- Name: staff_ratings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.staff_ratings (
    id integer NOT NULL,
    order_id integer,
    staff_id integer,
    staff_role character varying(50) NOT NULL,
    rating integer NOT NULL,
    comment text,
    created_at timestamp with time zone DEFAULT now(),
    CONSTRAINT staff_ratings_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


ALTER TABLE public.staff_ratings OWNER TO postgres;

--
-- Name: staff_ratings_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.staff_ratings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.staff_ratings_id_seq OWNER TO postgres;

--
-- Name: staff_ratings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.staff_ratings_id_seq OWNED BY public.staff_ratings.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    full_name character varying(255) NOT NULL,
    phone character varying(20) NOT NULL,
    password_hash text NOT NULL,
    role public.user_role DEFAULT 'customer'::public.user_role NOT NULL,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: categories id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);


--
-- Name: expenses id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.expenses ALTER COLUMN id SET DEFAULT nextval('public.expenses_id_seq'::regclass);


--
-- Name: ingredients id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ingredients ALTER COLUMN id SET DEFAULT nextval('public.ingredients_id_seq'::regclass);


--
-- Name: order_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items ALTER COLUMN id SET DEFAULT nextval('public.order_items_id_seq'::regclass);


--
-- Name: orders id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);


--
-- Name: product_ingredients id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_ingredients ALTER COLUMN id SET DEFAULT nextval('public.product_ingredients_id_seq'::regclass);


--
-- Name: products id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);


--
-- Name: staff_ratings id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_ratings ALTER COLUMN id SET DEFAULT nextval('public.staff_ratings_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: bot_staff; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.bot_staff (telegram_id, role, created_at) FROM stdin;
\.


--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.categories (id, name, image_url, is_user_controlled, created_at, updated_at) FROM stdin;
10	Milliy Taomlar	https://images.unsplash.com/photo-1529042410759-befb1204b468?auto=format&fit=crop&q=80&w=400	t	2026-06-24 14:45:26.608175-04	2026-06-24 14:45:26.608175-04
11	Fast Food	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	2026-06-24 14:45:26.608175-04	2026-06-24 14:45:26.608175-04
12	Ichimliklar	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	2026-06-24 14:45:26.608175-04	2026-06-24 14:45:26.608175-04
13	Salatlar	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	2026-06-24 14:45:26.608175-04	2026-06-24 14:45:26.608175-04
14	Shirinliklar	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	2026-06-24 14:45:26.608175-04	2026-06-24 14:45:26.608175-04
\.


--
-- Data for Name: expenses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.expenses (id, amount, category, description, created_at) FROM stdin;
\.


--
-- Data for Name: ingredients; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.ingredients (id, name, stock, unit, min_stock, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.order_items (id, order_id, product_id, quantity, price, unit, comment, created_at) FROM stdin;
1	1	\N	40.000	1000.00	dona		2026-06-24 14:25:32.029394-04
2	2	3	4.000	6250.00	dona		2026-06-24 14:49:03.871203-04
3	2	3	1.000	25000.00	pors		2026-06-24 14:49:03.871203-04
4	2	7	10.000	7000.00	dona		2026-06-24 14:49:03.871203-04
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.orders (id, customer_id, total_price, status, address, phone, lat, lng, courier_id, cook_id, comment, created_at, updated_at) FROM stdin;
1	1	40000.00	delivered	warwgdhfmgj yf,gh	+998886288822	\N	\N	1	\N		2026-06-24 14:25:32.029394-04	2026-06-24 14:26:05.682204-04
2	1	120000.00	delivered	wdwdwdwdwdw	+998886288822	\N	\N	1	\N		2026-06-24 14:49:03.871203-04	2026-06-24 14:50:13.749142-04
\.


--
-- Data for Name: product_ingredients; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.product_ingredients (id, product_id, ingredient_id, quantity, unit, created_at) FROM stdin;
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.products (id, category_id, name, description, price, image_url, is_active, unit, min_quantity, quantity_step, has_mandatory_container, created_at, updated_at) FROM stdin;
3	10	Osh (Palov)	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
4	10	Qozon kabob	Mijoz uchun namunaviy mahsulot	45000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
5	10	Lag'mon	Mijoz uchun namunaviy mahsulot	22000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
6	10	Manti	Mijoz uchun namunaviy mahsulot	4000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	dona	4.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
7	10	Somsa	Mijoz uchun namunaviy mahsulot	7000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
8	10	Sho'rva	Mijoz uchun namunaviy mahsulot	20000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
9	10	Shashlik (Qiyma)	Mijoz uchun namunaviy mahsulot	12000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	dona	2.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
10	10	Shashlik (Jaz)	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	dona	2.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
11	10	Norin	Mijoz uchun namunaviy mahsulot	30000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
12	10	Hasip	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
13	10	Qovurdoq	Mijoz uchun namunaviy mahsulot	40000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
14	10	Tuxum barak	Mijoz uchun namunaviy mahsulot	18000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
15	10	Dimlama	Mijoz uchun namunaviy mahsulot	35000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
16	10	Beshbarmoq	Mijoz uchun namunaviy mahsulot	50000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
17	10	Moshxo'rda	Mijoz uchun namunaviy mahsulot	18000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
18	10	Chuchvara	Mijoz uchun namunaviy mahsulot	20000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
19	10	Kavob	Mijoz uchun namunaviy mahsulot	35000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
20	10	Go'shtli non	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
21	10	Qo'y go'shti (Xom)	Mijoz uchun namunaviy mahsulot	95000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	kg	1.000	0.500	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
22	10	Mol go'shti (Xom)	Mijoz uchun namunaviy mahsulot	85000.00	https://images.unsplash.com/photo-1546069901-ba9599a7e63c?auto=format&fit=crop&q=80&w=400	t	kg	1.000	0.500	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
23	11	Burger	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
24	11	Cheeseburger	Mijoz uchun namunaviy mahsulot	28000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
25	11	Hot-dog	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
26	11	Klab Sendvich	Mijoz uchun namunaviy mahsulot	30000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
27	11	KFC Tovuq (Qanotcha)	Mijoz uchun namunaviy mahsulot	35000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
28	11	KFC Tovuq (Strips)	Mijoz uchun namunaviy mahsulot	32000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
29	11	Fri kartoshkasi	Mijoz uchun namunaviy mahsulot	12000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
30	11	Nonli hot-dog	Mijoz uchun namunaviy mahsulot	14000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
31	11	Pitsa (Margarita)	Mijoz uchun namunaviy mahsulot	55000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
32	11	Pitsa (Go'shtli)	Mijoz uchun namunaviy mahsulot	75000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
33	11	Pitsa (Qo'ziqorinli)	Mijoz uchun namunaviy mahsulot	65000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
34	11	Pitsa (Arlash)	Mijoz uchun namunaviy mahsulot	70000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
35	11	Donar Kebab	Mijoz uchun namunaviy mahsulot	28000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
36	11	Lavash (Go'shtli)	Mijoz uchun namunaviy mahsulot	26000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
37	11	Lavash (Tovuqli)	Mijoz uchun namunaviy mahsulot	24000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
38	11	Shaurma	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
39	11	Gamburger mini	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
40	11	Katta Combo	Mijoz uchun namunaviy mahsulot	60000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
41	11	Kichik Combo	Mijoz uchun namunaviy mahsulot	45000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
42	11	Qovurilgan Tovuq	Mijoz uchun namunaviy mahsulot	40000.00	https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
43	12	Coca-Cola (1.5L)	Mijoz uchun namunaviy mahsulot	14000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
44	12	Fanta (1.5L)	Mijoz uchun namunaviy mahsulot	14000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
45	12	Sprite (1.5L)	Mijoz uchun namunaviy mahsulot	14000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
46	12	Coca-Cola (0.5L)	Mijoz uchun namunaviy mahsulot	7000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
47	12	Fanta (0.5L)	Mijoz uchun namunaviy mahsulot	7000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
48	12	Choy (Qora)	Mijoz uchun namunaviy mahsulot	5000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
49	12	Choy (Ko'k)	Mijoz uchun namunaviy mahsulot	5000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
50	12	Limon choy	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
51	12	Kofe (Amerikano)	Mijoz uchun namunaviy mahsulot	12000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
52	12	Kofe (Latte)	Mijoz uchun namunaviy mahsulot	18000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
53	12	Kofe (Kapuchino)	Mijoz uchun namunaviy mahsulot	16000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
54	12	Sok (Olma 1L)	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
55	12	Sok (Gilos 1L)	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
56	12	Sok (O'rik 1L)	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
57	12	Suv (Gazlangan 1L)	Mijoz uchun namunaviy mahsulot	4000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
58	12	Suv (Gazsiz 1L)	Mijoz uchun namunaviy mahsulot	4000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
59	12	Moxito	Mijoz uchun namunaviy mahsulot	20000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
60	12	Ayron	Mijoz uchun namunaviy mahsulot	6000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
61	12	Qimiz	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
62	12	Qulupnayli sheyk	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1544145945-f90425340c7e?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
63	13	Achchiq-chuchuk	Mijoz uchun namunaviy mahsulot	12000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
64	13	Svejiy salat	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
65	13	Sezar	Mijoz uchun namunaviy mahsulot	35000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
66	13	Mujskoy kapriz	Mijoz uchun namunaviy mahsulot	32000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
67	13	Olivye	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
68	13	Grek salati	Mijoz uchun namunaviy mahsulot	28000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
69	13	Karam salati	Mijoz uchun namunaviy mahsulot	10000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
70	13	Sabzi salati (Koreyscha)	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
71	13	Qo'ziqorinli salat	Mijoz uchun namunaviy mahsulot	30000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
72	13	Tovuqli salat	Mijoz uchun namunaviy mahsulot	28000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
73	13	Vinigret	Mijoz uchun namunaviy mahsulot	20000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
74	13	Baxor salati	Mijoz uchun namunaviy mahsulot	18000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
75	13	Krabli salat	Mijoz uchun namunaviy mahsulot	22000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
76	13	Selyodka pod shuboy	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
77	13	Qaldirg'och uyasi	Mijoz uchun namunaviy mahsulot	30000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
78	13	Gullar salati	Mijoz uchun namunaviy mahsulot	24000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
79	13	Dungan salati	Mijoz uchun namunaviy mahsulot	20000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
80	13	Pishloqli salat	Mijoz uchun namunaviy mahsulot	26000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
81	13	Makkajo'xori salat	Mijoz uchun namunaviy mahsulot	18000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
82	13	Go'shtli assorti	Mijoz uchun namunaviy mahsulot	150000.00	https://images.unsplash.com/photo-1512621776951-a57141f2eefd?auto=format&fit=crop&q=80&w=400	t	kg	0.500	0.100	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
83	14	Asalli tort	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
84	14	Napaleon	Mijoz uchun namunaviy mahsulot	18000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
85	14	Snikers tort	Mijoz uchun namunaviy mahsulot	20000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
86	14	Muzqaymoq (Plombir)	Mijoz uchun namunaviy mahsulot	10000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
87	14	Muzqaymoq (Shokoladli)	Mijoz uchun namunaviy mahsulot	12000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
88	14	Muzqaymoq (Meva assorti)	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
89	14	Chizkeyk (Klassik)	Mijoz uchun namunaviy mahsulot	22000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
90	14	Chizkeyk (Qulupnayli)	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
91	14	Pechenye assorti	Mijoz uchun namunaviy mahsulot	40000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	kg	0.500	0.100	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
92	14	Shokoladli pechenye	Mijoz uchun namunaviy mahsulot	45000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	kg	0.500	0.100	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
93	14	Shakolatlar	Mijoz uchun namunaviy mahsulot	120000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	kg	0.100	0.100	f	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
94	14	Marmelad	Mijoz uchun namunaviy mahsulot	80000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	kg	0.200	0.100	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
95	14	Eksler (Krem) - 100gr	Mijoz uchun namunaviy mahsulot	8000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	gr	200.000	100.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
96	14	Makkaron (Pechenye)	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	dona	2.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
97	14	Keks	Mijoz uchun namunaviy mahsulot	12000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	dona	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
98	14	Rulet (Meva)	Mijoz uchun namunaviy mahsulot	18000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
99	14	Paxlava	Mijoz uchun namunaviy mahsulot	25000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
100	14	Chak-chak	Mijoz uchun namunaviy mahsulot	60000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	kg	0.300	0.100	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
101	14	Yong'oqli pishiriq	Mijoz uchun namunaviy mahsulot	85000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	kg	0.200	0.100	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
102	14	Xolvaytar	Mijoz uchun namunaviy mahsulot	15000.00	https://images.unsplash.com/photo-1551024506-0bccd828d307?auto=format&fit=crop&q=80&w=400	t	pors	1.000	1.000	t	2026-06-24 14:45:58.270227-04	2026-06-24 14:45:58.270227-04
\.


--
-- Data for Name: settings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.settings (key, value, updated_at) FROM stdin;
container_price	1000	2026-06-24 14:27:06.523358-04
container_product_id	7	2026-06-24 14:27:06.530381-04
\.


--
-- Data for Name: staff_ratings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.staff_ratings (id, order_id, staff_id, staff_role, rating, comment, created_at) FROM stdin;
1	1	1	courier	5	wdwdw	2026-06-24 14:32:50.593205-04
2	2	1	courier	5	wdwdw	2026-06-24 14:54:57.75021-04
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, full_name, phone, password_hash, role, created_at, updated_at) FROM stdin;
1	Admin	+998886288822	$2a$14$BrsE22ue4sG332iAbwuIQ.gbwAArzDykn0urSAC4CVDLMpZMwSQ6e	admin	2026-06-24 14:20:45.945774-04	2026-06-24 14:20:45.945774-04
2	Oshpaz	+998913338228	$2a$14$tDP4x72GlGAxvxpspgV0IuYOoZADqa1a5Lh1CHDWVA7ErZWncWtYK	cook	2026-06-24 14:50:53.978483-04	2026-06-24 14:50:53.978483-04
\.


--
-- Name: categories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.categories_id_seq', 1, true);


--
-- Name: expenses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.expenses_id_seq', 1, false);


--
-- Name: ingredients_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ingredients_id_seq', 1, false);


--
-- Name: order_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.order_items_id_seq', 4, true);


--
-- Name: orders_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.orders_id_seq', 2, true);


--
-- Name: product_ingredients_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.product_ingredients_id_seq', 1, false);


--
-- Name: products_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.products_id_seq', 102, true);


--
-- Name: staff_ratings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.staff_ratings_id_seq', 5, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 2, true);


--
-- Name: bot_staff bot_staff_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.bot_staff
    ADD CONSTRAINT bot_staff_pkey PRIMARY KEY (telegram_id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: expenses expenses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_pkey PRIMARY KEY (id);


--
-- Name: ingredients ingredients_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ingredients
    ADD CONSTRAINT ingredients_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: product_ingredients product_ingredients_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_ingredients
    ADD CONSTRAINT product_ingredients_pkey PRIMARY KEY (id);


--
-- Name: product_ingredients product_ingredients_product_id_ingredient_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_ingredients
    ADD CONSTRAINT product_ingredients_product_id_ingredient_id_key UNIQUE (product_id, ingredient_id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: settings settings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settings
    ADD CONSTRAINT settings_pkey PRIMARY KEY (key);


--
-- Name: staff_ratings staff_ratings_order_id_staff_id_staff_role_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_ratings
    ADD CONSTRAINT staff_ratings_order_id_staff_id_staff_role_key UNIQUE (order_id, staff_id, staff_role);


--
-- Name: staff_ratings staff_ratings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_ratings
    ADD CONSTRAINT staff_ratings_pkey PRIMARY KEY (id);


--
-- Name: users users_phone_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_phone_key UNIQUE (phone);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: categories update_categories_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON public.categories FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: orders update_orders_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON public.orders FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: products update_products_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE SET NULL;


--
-- Name: orders orders_cook_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_cook_id_fkey FOREIGN KEY (cook_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: orders orders_courier_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_courier_id_fkey FOREIGN KEY (courier_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: orders orders_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: product_ingredients product_ingredients_ingredient_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_ingredients
    ADD CONSTRAINT product_ingredients_ingredient_id_fkey FOREIGN KEY (ingredient_id) REFERENCES public.ingredients(id) ON DELETE CASCADE;


--
-- Name: product_ingredients product_ingredients_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.product_ingredients
    ADD CONSTRAINT product_ingredients_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: products products_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id) ON DELETE CASCADE;


--
-- Name: staff_ratings staff_ratings_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_ratings
    ADD CONSTRAINT staff_ratings_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: staff_ratings staff_ratings_staff_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.staff_ratings
    ADD CONSTRAINT staff_ratings_staff_id_fkey FOREIGN KEY (staff_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict cRrSX3qdpMuxIgMUQZSOvh0it6b0DwLHkpnBnoV6iKMcSNGwSxcFLfuSTYWIkbE

