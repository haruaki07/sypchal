-- +goose Up
-- +goose StatementBegin
ALTER TABLE "orders" ADD COLUMN "pay_id" VARCHAR NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "orders" DROP COLUMN "pay_id";
-- +goose StatementEnd
