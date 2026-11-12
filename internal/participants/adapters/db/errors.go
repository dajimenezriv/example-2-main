package db

const uniqueViolation = "23505"

// Constraint names on the actors table (see migrations/).
const (
	actorsPKeyConstraint  = "actors_pkey"
	actorsTaxIDConstraint = "actors_tax_id_key"
)
