-- name: CreateActor :one
INSERT INTO
  actors (id, tax_id, type)
VALUES
  (@id, @tax_id, @type) RETURNING *;

-- name: FindActorByID :one
SELECT
  *
FROM
  actors
WHERE
  id = @id;

-- name: FindActors :many
SELECT
  *
FROM
  actors;