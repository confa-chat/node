-- +goose Up
-- +goose StatementBegin
-- User Table
CREATE TABLE "user" (
    id uuid PRIMARY KEY,
    username VARCHAR(255) UNIQUE
);
-- Server Table
CREATE TABLE "server" (id uuid PRIMARY KEY, "name" VARCHAR(255));
-- TextChannel Table
CREATE TABLE "text_channel" (
    id uuid PRIMARY KEY,
    server_id uuid,
    FOREIGN KEY (server_id) REFERENCES "server"(id),
    "name" VARCHAR(255)
);
CREATE UNIQUE INDEX unique_text_channel_name ON "text_channel"(server_id, "name");
-- VoiceChannel Table
CREATE TABLE "voice_channel" (
    id uuid PRIMARY KEY,
    server_id uuid,
    FOREIGN KEY (server_id) REFERENCES "server"(id),
    "name" VARCHAR(255)
);
-- Message Table
CREATE TABLE "message" (
    id uuid PRIMARY KEY,
    channel_id uuid NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES "text_channel"(id),
    sender_id uuid NOT NULL,
    FOREIGN KEY (sender_id) REFERENCES "user"(id),
    content TEXT NOT NULL,
    "timestamp" TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX message_channel ON "message" USING hash ("channel_id");
CREATE INDEX message_timestamp ON "message" USING brin ("timestamp");
-- +goose StatementEnd