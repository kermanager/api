-- Drop tables
DROP TABLE IF EXISTS "tickets";
DROP TABLE IF EXISTS "tombolas";
DROP TABLE IF EXISTS "participations";
DROP TABLE IF EXISTS "kermesses_stands";
DROP TABLE IF EXISTS "kermesses_users";
DROP TABLE IF EXISTS "kermesses";
DROP TABLE IF EXISTS "stands";
DROP TABLE IF EXISTS "users";

-- Drop custom types
DROP TYPE IF EXISTS tombolas_status_enum;
DROP TYPE IF EXISTS stands_type_enum;
DROP TYPE IF EXISTS users_role_enum;