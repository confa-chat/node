-- +goose Up
-- +goose StatementBegin
-- User Table
CREATE TABLE "user" (
    id uuid PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE
);
-- Server Table
CREATE TABLE "server" (id uuid PRIMARY KEY, "name" VARCHAR(255));
-- TextChannel Table
CREATE TABLE "text_channel" (
    id UUID PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES "server"(id) ON DELETE CASCADE,
    "name" VARCHAR(255) NOT NULL
);
CREATE UNIQUE INDEX unique_text_channel_name ON "text_channel"(server_id, "name");
-- VoiceChannel Table
CREATE TABLE "voice_channel" (
    id uuid PRIMARY KEY,
    server_id uuid NOT NULL REFERENCES "server"(id) ON DELETE CASCADE,
    "name" VARCHAR(255) NOT NULL
);
-- Message Table
CREATE TABLE "message" (
    id uuid PRIMARY KEY,
    channel_id uuid NOT NULL REFERENCES "text_channel"(id) ON DELETE CASCADE,
    sender_id uuid NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    "timestamp" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX message_channel ON "message" USING hash ("channel_id");
CREATE INDEX message_timestamp ON "message" USING brin ("timestamp");
-- +goose StatementEnd