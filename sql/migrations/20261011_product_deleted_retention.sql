begin;

alter table public.products
  add column if not exists deleted_at timestamp with time zone;

alter table public.products
  drop constraint if exists products_status_check;

alter table public.products
  add constraint products_status_check
  check (status in ('active', 'inactive', 'sold', 'deleted'));

create index if not exists idx_products_deleted_retention
  on public.products (deleted_at)
  where status = 'deleted' and deleted_at is not null;

commit;
