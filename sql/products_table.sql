create table public.products (
  id uuid not null default gen_random_uuid (),
  name text not null,
  sku text not null,
  description text not null,
  price bigint not null,
  discount bigint not null default 0,
  tax bigint not null default 0,
  subtotal bigint not null default 0,
  stock integer not null default 0,
  is_active boolean not null default true,
  created_at timestamp with time zone not null default now(),
  updated_at timestamp with time zone not null default now(),
  product_type text null,
  constraint products_pkey primary key (id),
  constraint products_sku_key unique (sku),
  constraint products_price_check check ((price >= 0)),
  constraint products_stock_check check ((stock >= 0)),
  constraint products_subtotal_check check ((subtotal >= 0)),
  constraint check_price_calculation check ((subtotal = ((price - discount) + tax))),
  constraint products_tax_check check ((tax >= 0)),
  constraint products_discount_check check ((discount >= 0))
) TABLESPACE pg_default;

create index IF not exists idx_products_active_created on public.products using btree (is_active, created_at desc) TABLESPACE pg_default;

create index IF not exists idx_products_active_only on public.products using btree (created_at desc, id) TABLESPACE pg_default
where
  (is_active = true);

create index IF not exists idx_products_inventory on public.products using btree (stock, is_active) TABLESPACE pg_default
where
  (is_active = true);

create index IF not exists idx_products_listing_cover on public.products using btree (is_active, created_at desc) INCLUDE (id, name, sku, price, description, stock) TABLESPACE pg_default
where
  (is_active = true);

create index IF not exists idx_products_name on public.products using btree (name) TABLESPACE pg_default;

create index IF not exists idx_products_price on public.products using btree (price) TABLESPACE pg_default
where
  (is_active = true);

create index IF not exists idx_products_search on public.products using gin (
  to_tsvector(
    'english'::regconfig,
    (
      (((name || ' '::text) || description) || ' '::text) || sku
    )
  )
) TABLESPACE pg_default;

create index IF not exists idx_products_sku on public.products using btree (sku) TABLESPACE pg_default;

create index IF not exists products_is_active_idx on public.products using btree (is_active) TABLESPACE pg_default;

create index IF not exists products_product_type_idx on public.products using btree (product_type) TABLESPACE pg_default;

create trigger set_updated_at BEFORE
update on products for EACH row
execute FUNCTION update_updated_at_column ();

---
create table public.product_images (
  id uuid not null default gen_random_uuid (),
  product_id uuid not null,
  url text null,
  alt_text text null,
  is_primary boolean not null default false,
  name text null,
  "position" integer not null default 0,
  constraint product_images_pkey primary key (id),
  constraint product_images_product_id_fkey foreign KEY (product_id) references products (id) on delete CASCADE
) TABLESPACE pg_default;

create index IF not exists idx_product_images_primary_only on public.product_images using btree (product_id) TABLESPACE pg_default
where
  (is_primary = true);

create index IF not exists idx_product_images_product_id on public.product_images using btree (product_id) TABLESPACE pg_default;

create index IF not exists idx_product_images_product_primary on public.product_images using btree (product_id, is_primary desc) TABLESPACE pg_default;

create unique INDEX IF not exists product_images_name_key on public.product_images using btree (name) TABLESPACE pg_default
where
  (name is not null);

create index IF not exists product_images_name_idx on public.product_images using btree (name) TABLESPACE pg_default;

create index IF not exists product_images_position_idx on public.product_images using btree ("position") TABLESPACE pg_default;

create index IF not exists product_images_is_primary_idx on public.product_images using btree (is_primary) TABLESPACE pg_default;

create index IF not exists product_images_alt_text_idx on public.product_images using btree (alt_text) TABLESPACE pg_default;

create trigger ensure_primary_image BEFORE INSERT
or
update on product_images for EACH row when (new.is_primary = true)
execute FUNCTION ensure_single_primary_image ();
