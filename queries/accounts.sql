-- name: CreateAccount :one
INSERT INTO
  accounts (
    id,
    actor_id,
    dlt_account_id,
    private_key,
    public_key
  )
VALUES
  (
    @id,
    @actor_id,
    @dlt_account_id,
    @private_key,
    @public_key
  ) RETURNING *;

-- name: FindAccountByID :one
SELECT
  *
FROM
  accounts
WHERE
  id = @id;

-- name: FindAccounts :many
SELECT
  *
FROM
  accounts;

-- name: FindAccountsByActorIDs :many
SELECT
  *
FROM
  accounts
WHERE
  actor_id = ANY(@actor_ids::uuid[]);