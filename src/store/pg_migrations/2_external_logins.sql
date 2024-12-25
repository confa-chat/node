-- +goose Up
-- +goose StatementBegin
CREATE TABLE "external_login" (
    "id" uuid PRIMARY KEY,
    "user_id" uuid NOT NULL REFERENCES "user" (id) ON DELETE CASCADE,
    "issuer" TEXT NOT NULL,
    "subject" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX external_login_unique_external_id ON "external_login" ("issuer", "subject");
-- +goose StatementEnd