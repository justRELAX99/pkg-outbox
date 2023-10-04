-- +goose Up
CREATE TABLE IF NOT EXISTS outbox(
    uuid uuid NOT NULL,
    message jsonb NULL DEFAULT '{}'::jsonb,
    state smallint NOT NULL,
    created_on int8 NOT NULL,
    service_name text NOT NULL,
    PRIMARY KEY (uuid)
);

-- +goose Down
DROP TABLE IF EXISTS outbox;