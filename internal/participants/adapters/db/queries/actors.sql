-- name: CreateActor :exec
INSERT INTO
  actors (id, tax_id, type)
VALUES
  (@id, @tax_id, @type);

-- name: UpdateActor :exec
UPDATE
  actors
SET
  tax_id = @tax_id,
  type = @type
WHERE
  id = @id;

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