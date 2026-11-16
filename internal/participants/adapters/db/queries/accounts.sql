-- name: CountAccounts :one
SELECT COUNT(*) FROM accounts;


-- name: CreateAccount :exec
INSERT INTO
  accounts (
    id,
    actor_id,
    dlt_account_id,
    private_key,
    public_key,
    version
  )
VALUES
  (
    @id,
    @actor_id,
    @dlt_account_id,
    @private_key,
    @public_key,
    @version
  );

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
  accounts
LIMIT
  @page_limit OFFSET @page_offset;

-- name: FindAccountsByActorIDs :many
SELECT
  *
FROM
  accounts
WHERE
  actor_id = ANY(@actor_ids::uuid[]);