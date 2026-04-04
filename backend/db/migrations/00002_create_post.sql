-- +goose Up
-- +goose StatementBegin
CREATE TABLE "post" (
  "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  "title" text NOT NULL,
  "content" text NOT NULL,
  "author_id" UUID NOT NULL REFERENCES "user"("id"),
  "created_at" timestamp NOT NULL DEFAULT now(),
  "updated_at" timestamp NOT NULL DEFAULT now()
);

CREATE INDEX ON "post" ("author_id");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "post";
-- +goose StatementEnd
