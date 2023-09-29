BEGIN;

create table twa.bridge_subscriptions
(
    id bigserial
        constraint bridge_subscription_pkey
            primary key,
    client_id        text not null,
    telegram_user_id bigint not null,
    origin           text not null,
    created_at       timestamp default now() not null,

    constraint unique_origin unique (telegram_user_id, origin)
);

COMMIT;
