-- name: CountActors :one
SELECT COUNT(*) FROM actors;

-- name: CreateActor :exec
INSERT INTO
  actors (id, tax_id, type, version)
VALUES
  (@id, @tax_id, @type, @version);

-- name: UpdateActor :one
UPDATE
  actors
SET
  tax_id = @tax_id,
  type = @type,
  version = version + 1
WHERE
  id = @id
  AND version = @version RETURNING *;

-- name: GetActorVersion :one
SELECT
  version
FROM
  actors
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
  actors
ORDER BY
  id
LIMIT
  @page_limit OFFSET @page_offset;

-- name: DeleteActor :execrows
DELETE FROM actors
WHERE
  id = @id
  AND version = @version;
