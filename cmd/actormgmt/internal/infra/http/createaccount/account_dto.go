package createaccount

type CreateAccountRequest struct {
	ActorID      string `json:"actor_id"`
	DltAccountId string `json:"dlt_account_id"`
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
}

type CreateAccountResponse struct {
	ID           string `json:"id"`
	ActorID      string `json:"actor_id"`
	DltAccountId string `json:"dlt_account_id"`
	PublicKey    string `json:"public_key"`
}
