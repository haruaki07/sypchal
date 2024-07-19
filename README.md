## Challange

### Usage

```shell
$ make # watch and run on development environment
$ make build # build an executable binary

$ make psql # connect to database with psql

$ make down # stop docker containers
```

### Endpoints

```shell
POST /api/register # register an account for customer
POST /api/login # customer login

POST /api/products # admin only, create products
PUT /api/products/:id # admin only, update products
DELETE /api/products/:id # admin only, delete products
GET /api/products # list all products
GET /api/products/:id # get product by id
GET /api/category/:category # get all products by category

POST /api/cart # add product(s) to cart, should update qty if already exists on cart
GET /api/cart # list all shopping cart items
DELETE /api/cart/:id # delete cart item by item id
PUT /api/cart/:id # update cart item quantity by item id

POST /api/order # place an order
POST /api/order/pay/:id # pay an order
```

### ERD

- DBML: [erd.dbml](erd.dbml)
- Preview: https://dbdocs.io/sharylolive/sypchal?view=relationships
