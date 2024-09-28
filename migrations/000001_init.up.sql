-- Table: users

CREATE TYPE users_role_enum AS ENUM ('MANAGER', 'STAND_HOLDER', 'PARENT', 'CHILD');

CREATE TABLE "users" (
  "id" SERIAL PRIMARY KEY,
  "name" VARCHAR(255) NOT NULL,
  "email" VARCHAR(255) UNIQUE NOT NULL,
  "password" VARCHAR(255) NOT NULL,
  "role" users_role_enum NOT NULL,
  "credit" INTEGER NOT NULL DEFAULT 0
  -- "code" VARCHAR(255) NOT NULL DEFAULT '',
  -- "code_expires_at" TIMESTAMP,
);

--- Table: Stands

CREATE TABLE "stands" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- stand holder
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT DEFAULT ''
);

--- Table: kermesses

CREATE TABLE "kermesses" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- manager
  "name" VARCHAR(255) NOT NULL,
  "description" TEXT DEFAULT ''
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

--- Table: Tombolas

CREATE TYPE tombolas_status_enum AS ENUM ('CREATED', 'STARTED', 'ENDED');

CREATE TABLE "tombolas" (
  "id" SERIAL PRIMARY KEY,
  "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
  "name" VARCHAR(255) NOT NULL,
  "status" VARCHAR(255) NOT NULL DEFAULT 'CREATED'
);

CREATE TABLE "tombolas_users" (
  "id" SERIAL PRIMARY KEY,
  "tombola_id" INTEGER NOT NULL REFERENCES "tombolas"("id"),
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- child
  UNIQUE ("tombola_id", "user_id")
);

--- Table: Participations

CREATE TABLE "participations" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id"), -- child / parent
  "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
  "stand_id" INTEGER NOT NULL REFERENCES "stands"("id"),
  "credit" INTEGER NOT NULL DEFAULT 0,
  "point" INTEGER NOT NULL DEFAULT 0
);