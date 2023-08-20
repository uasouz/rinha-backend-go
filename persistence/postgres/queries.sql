-- name: AddPerson :one
insert into people (uuid,name,nickname,birthdate,stack)
    values ($1,$2,$3,$4,$5) RETURNING id;

-- name: GetPerson :one
SELECT id,uuid,name,nickname,birthdate,stack,created_at
    FROM people WHERE uuid = $1;

-- name: GetPeople :many
select id,uuid,name,nickname,birthdate,stack,created_at
    from people;

-- name: CountPeople :one
select count(*) from people;