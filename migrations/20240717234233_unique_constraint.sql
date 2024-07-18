-- +goose Up
-- +goose StatementBegin
ALTER TABLE cart_items ADD CONSTRAINT cart_items_user_id_product_id_key UNIQUE (user_id, product_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cart_items DROP CONSTRAINT cart_items_user_id_product_id_key;
-- +goose StatementEnd
