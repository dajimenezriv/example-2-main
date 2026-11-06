package createaccount

type CreateAccountRequest struct {
	DltAccountId string `json:"dlt_account_id"`
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
}

type CreateAccountResponse struct {
	ID           string `json:"id"`
	DltAccountId string `json:"dlt_account_id"`
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
}
