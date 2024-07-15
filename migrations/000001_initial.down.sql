BEGIN;

drop table "order_items";
drop table "cart_items";
drop table "orders";
drop table "products";
drop table "payments";
drop table "users";
drop type "order_status";

COMMIT;