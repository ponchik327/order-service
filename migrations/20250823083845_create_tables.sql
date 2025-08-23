-- +goose Up
-- Создание таблицы orders
CREATE TABLE orders (
    order_uid VARCHAR(50) PRIMARY KEY,
    track_number VARCHAR(50),
    entry VARCHAR(50),
    locale VARCHAR(10),
    internal_signature VARCHAR(100),
    customer_id VARCHAR(50),
    delivery_service VARCHAR(50),
    shardkey VARCHAR(10),
    sm_id INTEGER,
    date_created TIMESTAMP,
    oof_shard VARCHAR(10)
);

-- Создание таблицы deliveries
CREATE TABLE deliveries (
    order_uid VARCHAR(50) PRIMARY KEY REFERENCES orders(order_uid),
    name VARCHAR(100),
    phone VARCHAR(20),
    zip VARCHAR(20),
    city VARCHAR(100),
    address VARCHAR(200),
    region VARCHAR(100),
    email VARCHAR(100)
);

-- Создание таблицы payments
CREATE TABLE payments (
    order_uid VARCHAR(50) PRIMARY KEY REFERENCES orders(order_uid),
    transaction VARCHAR(50),
    request_id VARCHAR(50),
    currency VARCHAR(10),
    provider VARCHAR(50),
    amount INTEGER,
    payment_dt BIGINT,
    bank VARCHAR(50),
    delivery_cost INTEGER,
    goods_total INTEGER,
    custom_fee INTEGER
);

-- Создание таблицы items
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(50) REFERENCES orders(order_uid),
    chrt_id INTEGER,
    track_number VARCHAR(50),
    price INTEGER,
    rid VARCHAR(50),
    name VARCHAR(100),
    sale INTEGER,
    size VARCHAR(10),
    total_price INTEGER,
    nm_id INTEGER,
    brand VARCHAR(100),
    status INTEGER
);

-- +goose Down
-- Откат миграции: удаление таблиц в обратном порядке
DROP TABLE items;
DROP TABLE payments;
DROP TABLE deliveries;
DROP TABLE orders;
