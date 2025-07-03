create type order_status as enum ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

create type order_type as enum ('BALANCE', 'WITHDRAW');

create table orders (id uuid primary key, 
    profile_id uuid not null, 
    order_num bigint not null,
    status order_status not null,
    uploaded_at timestamp with time zone not null,
    sum numeric(10, 2),
    type_order order_type not null,
    CONSTRAINT order_num_unique UNIQUE(order_num),
    CONSTRAINT fk_profile foreign key (profile_id) references profile (id));