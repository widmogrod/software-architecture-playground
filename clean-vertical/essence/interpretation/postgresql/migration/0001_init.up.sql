create table if not exists user_identity
(
    uuid          varchar unique,
    email_address varchar unique
);

create table if not exists  activation_tokens
(
    uuid  varchar unique,
    token varchar unique
);