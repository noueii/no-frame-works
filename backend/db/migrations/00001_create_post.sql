-- +goose Up
CREATE TABLE "post" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "title" text NOT NULL,
  "content" text NOT NULL,
  "author_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now()
);
CREATE INDEX ON "post" ("author_id");

-- +goose Down
DROP TABLE IF EXISTS "post";
