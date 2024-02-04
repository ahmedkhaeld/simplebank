CREATE TABLE "users" (
    "username" varchar(255)  PRIMARY KEY,
    "password" varchar(255) NOT NULL,
    "email" varchar(255) NOT NULL UNIQUE,
    "full_name" varchar(255) NOT NULL,
    "password_changed_at" timestamp NOT NULL DEFAULT '0001-01-01 00:00:00Z',
    "created_at" timestamp NOT NULL DEFAULT (now())
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");



-- create a composite key for owner and currency
-- make sure no duplicate currencies for the same account
ALTER TABLE "accounts" ADD CONSTRAINT "owner_currency_key"  UNIQUE ("owner", "currency");
