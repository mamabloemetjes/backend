--
-- PostgreSQL database dump
--


-- Dumped from database version 17.11 (Debian 17.11-1.pgdg13+2)
-- Dumped by pg_dump version 17.11 (Debian 17.11-1.pgdg13+2)

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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: product_images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    alt_text text,
    is_primary boolean DEFAULT false NOT NULL,
    name text,
    "position" integer DEFAULT 0 NOT NULL
);


--
-- Name: TABLE product_images; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.product_images IS 'Product images with support for multiple images per product';


--
-- Name: COLUMN product_images.is_primary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.product_images.is_primary IS 'Indicates if this is the primary/featured image for the product';


--
-- Name: products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    sku text NOT NULL,
    description text NOT NULL,
    price bigint NOT NULL,
    discount bigint DEFAULT 0 NOT NULL,
    tax bigint DEFAULT 0 NOT NULL,
    subtotal bigint DEFAULT 0 NOT NULL,
    stock integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'active'::text NOT NULL,
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    product_type text,
    CONSTRAINT check_price_calculation CHECK ((subtotal = ((price - discount) + tax))),
    CONSTRAINT products_discount_check CHECK ((discount >= 0)),
    CONSTRAINT products_price_check CHECK ((price >= 0)),
    CONSTRAINT products_status_check CHECK ((status = ANY (ARRAY['active'::text, 'inactive'::text, 'sold'::text, 'deleted'::text]))),
    CONSTRAINT products_stock_check CHECK ((stock >= 0)),
    CONSTRAINT products_subtotal_check CHECK ((subtotal >= 0)),
    CONSTRAINT products_tax_check CHECK ((tax >= 0))
);


--
-- Name: TABLE products; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON TABLE public.products IS 'Main products table storing all product information including flowers and bouquets';


--
-- Name: COLUMN products.sku; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.products.sku IS 'Stock Keeping Unit - unique identifier for inventory management';


--
-- Name: COLUMN products.price; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.products.price IS 'Product price stored in cents for precision (e.g., $19.99 = 1999)';


--
-- Name: COLUMN products.discount; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.products.discount IS 'Discount amount in cents';


--
-- Name: COLUMN products.tax; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.products.tax IS 'Tax amount in cents';


--
-- Name: COLUMN products.subtotal; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.products.subtotal IS 'Subtotal in cents (price - discount + tax)';


--
-- Name: COLUMN products.stock; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.products.stock IS 'Current inventory stock level';


--
-- Name: COLUMN products.product_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.products.product_type IS 'the product type';


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: products products_sku_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_sku_key UNIQUE (sku);


--
-- Name: idx_product_images_primary_only; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_images_primary_only ON public.product_images USING btree (product_id) WHERE (is_primary = true);


--
-- Name: idx_product_images_product_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_images_product_id ON public.product_images USING btree (product_id);


--
-- Name: idx_product_images_product_primary; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_product_images_product_primary ON public.product_images USING btree (product_id, is_primary DESC);


--
-- Name: idx_products_inventory; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_inventory ON public.products USING btree (stock) WHERE (status = 'active'::text);


--
-- Name: idx_products_listing_cover; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_listing_cover ON public.products USING btree (created_at DESC, id) INCLUDE (name, sku, price, stock) WHERE (status = 'active'::text);


--
-- Name: idx_products_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_name ON public.products USING btree (name);


--
-- Name: idx_products_search; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_search ON public.products USING gin (to_tsvector('english'::regconfig, ((((name || ' '::text) || description) || ' '::text) || sku)));


--
-- Name: product_images_alt_text_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX product_images_alt_text_idx ON public.product_images USING btree (alt_text);


--
-- Name: product_images_is_primary_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX product_images_is_primary_idx ON public.product_images USING btree (is_primary);


--
-- Name: product_images_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX product_images_name_idx ON public.product_images USING btree (name);


--
-- Name: product_images_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX product_images_name_key ON public.product_images USING btree (name) WHERE (name IS NOT NULL);


--
-- Name: product_images_position_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX product_images_position_idx ON public.product_images USING btree ("position");


--
-- Name: products_product_type_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX products_product_type_idx ON public.products USING btree (product_type);


--
-- Name: product_images ensure_primary_image; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER ensure_primary_image BEFORE INSERT OR UPDATE ON public.product_images FOR EACH ROW WHEN ((new.is_primary = true)) EXECUTE FUNCTION public.ensure_single_primary_image();


--
-- Name: products set_updated_at; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_updated_at BEFORE UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: product_images product_images_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;

--
-- PostgreSQL database dump complete
--
