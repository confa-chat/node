-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS message_attachment (
    id UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES message(id),
    name TEXT NOT NULL,
    attachment_id UUID NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_message_attachment_message_id ON message_attachment(message_id);
-- +goose StatementEnd