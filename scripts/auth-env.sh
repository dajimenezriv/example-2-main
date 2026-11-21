#!/usr/bin/env bash
set -euo pipefail

touch .env

upsert() {
  v="$(terraform -chdir=infra/auth0 output -raw "$2")" # plain assignment so set -e catches failures
  K="$1" V="$v" awk '
    $0 ~ "^" ENVIRON["K"] "=" { print ENVIRON["K"] "=\"" ENVIRON["V"] "\""; done = 1; next }
    { print }
    END { if (!done) print ENVIRON["K"] "=\"" ENVIRON["V"] "\"" }
  ' .env > .env.tmp
  cat .env.tmp > .env
  rm .env.tmp
}

upsert AUTH_ISSUER auth_issuer
upsert AUTH_AUDIENCE auth_audience
upsert CLI_AUTH0_CLIENT_ID cli_client_id
echo "Updated .env from Terraform outputs."
