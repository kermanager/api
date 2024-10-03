-- Table: users

CREATE TYPE users_role_enum AS ENUM ('MANAGER', 'STAND_HOLDER', 'PARENT', 'CHILD');

CREATE TABLE "users" (
  "id" SERIAL PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "email" VARCHAR(255) UNIQUE NOT NULL,
  "password" VARCHAR(255) NOT NULL,
  "role" users_role_enum NOT NULL,
  "credit" INTEGER NOT NULL DEFAULT 0,
  "parent_id" INTEGER REFERENCES "users"("id") DEFAULT NULL -- parent
);

--- Table: Stands

CREATE TYPE stands_type_enum AS ENUM ('CONSUMPTION', 'ACTIVITY');

CREATE TABLE "stands" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER NOT NULL UNIQUE REFERENCES "users"("id"), -- stand holder
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT DEFAULT '',
  "type" stands_type_enum NOT NULL,
  "price" INTEGER NOT NULL DEFAULT 0,
  "stock" INTEGER NOT NULL DEFAULT 0
);

--- Table: kermesses

CREATE TYPE kermesses_status_enum AS ENUM ('STARTED', 'ENDED');

CREATE TABLE "kermesses" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- manager
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT DEFAULT '',
  "status" kermesses_status_enum NOT NULL DEFAULT 'STARTED'
);

CREATE TABLE "kermesses_users" (
  "id" SERIAL PRIMARY KEY,
  "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- child / parent
  UNIQUE ("kermesse_id", "user_id")
);

CREATE TABLE "kermesses_stands" (
  "id" SERIAL PRIMARY KEY,
  "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
  "stand_id" INTEGER NOT NULL REFERENCES "stands"("id"),
  UNIQUE ("kermesse_id", "stand_id")
);

--- Table: Interactions

CREATE TYPE interactions_type_enum AS ENUM ('CONSUMPTION', 'ACTIVITY');
CREATE TYPE interactions_status_enum AS ENUM ('STARTED', 'ENDED');

CREATE TABLE "interactions" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- child / parent
  "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
  "stand_id" INTEGER NOT NULL REFERENCES "stands"("id"),
  "type" interactions_type_enum NOT NULL,
  "status" interactions_status_enum NOT NULL DEFAULT 'STARTED',
  "credit" INTEGER NOT NULL DEFAULT 0,
  "point" INTEGER NOT NULL DEFAULT 0
);

--- Table: Tombolas

CREATE TYPE tombolas_status_enum AS ENUM ('STARTED', 'ENDED');

CREATE TABLE "tombolas" (
  "id" SERIAL PRIMARY KEY,
  "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
  "name" VARCHAR(255) NOT NULL,
  "status" tombolas_status_enum NOT NULL DEFAULT 'STARTED',
  "price" INTEGER NOT NULL DEFAULT 0,
  "gift" VARCHAR(255) NOT NULL
);

--- Table: Tickets

CREATE TABLE "tickets" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- child
  "tombola_id" INTEGER NOT NULL REFERENCES "tombolas"("id"),
  "is_winner" BOOLEAN NOT NULL DEFAULT FALSE
);
