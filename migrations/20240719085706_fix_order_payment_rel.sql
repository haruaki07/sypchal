-- +goose Up
-- +goose StatementBegin
ALTER TABLE "orders" DROP CONSTRAINT orders_id_fkey;
ALTER TABLE "payments" ADD FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "orders" ADD FOREIGN KEY ("id") REFERENCES "payments" ("order_id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "orders" DROP CONSTRAINT payments_order_id_fkey;
-- +goose StatementEnd
