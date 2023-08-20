-- migrate:up
    CREATE TABLE people (
      id serial PRIMARY KEY,
      uuid uuid not null default gen_random_uuid(),
      name varchar(200) not null,
      nickname varchar(64) not null,
      birthdate DATE not null,
      stack varchar(64)[] not null,
      created_at  timestamp default current_timestamp
    );

    CREATE UNIQUE INDEX IF NOT EXISTS people_uuid_idx ON people (uuid);
    CREATE INDEX IF NOT EXISTS people_name_idx ON people (name);
    CREATE INDEX IF NOT EXISTS people_nickname_idx ON people (nickname);
    CREATE INDEX IF NOT EXISTS people_birthdate_idx ON people (birthdate);
    CREATE INDEX IF NOT EXISTS people_stack_idx ON people USING GIN (stack);
-- migrate:down

DROP TABLE IF EXISTS people;

