-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "email" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX ON "user" ("email");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "user";
-- +goose StatementEnd
