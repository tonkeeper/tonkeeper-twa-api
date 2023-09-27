BEGIN;
create schema if not exists twa;

create table twa.subscriptions
(
    id bigserial
        constraint subscription_pkey
            primary key,
    account          text not null,
    telegram_user_id bigint not null,
    created_at       timestamp default now() not null,

    constraint unique_account_and_user_id unique (account, telegram_user_id)
);

COMMIT;
