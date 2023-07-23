package tron

type RawTransaction struct {
	Visible    bool    `json:"visible"`
	TxID       string  `json:"txID"`
	RawData    RawData `json:"raw_data"`
	RawDataHex string  `json:"raw_data_hex"`
}

type SignedTransaction struct {
	Visible    bool     `json:"visible"`
	TxID       string   `json:"txID"`
	RawData    string   `json:"raw_data"`
	RawDataHex string   `json:"raw_data_hex"`
	Signature  []string `json:"signature,omitempty"`
}

type RawData struct {
	Contract []struct {
		Parameter struct {
			Value struct {
				Data            string `json:"data"`
				Amount          int    `json:"amount"`
				OwnerAddress    string `json:"owner_address"`
				ContractAddress string `json:"contract_address"`
				ToAddress       string `json:"to_address"`
			} `json:"value"`
			TypeURL string `json:"type_url"`
		} `json:"parameter"`
		Type string `json:"type"`
	} `json:"contract"`
	RefBlockBytes string `json:"ref_block_bytes"`
	RefBlockHash  string `json:"ref_block_hash"`
	Expiration    int64  `json:"expiration"`
	Timestamp     int64  `json:"timestamp"`
	FeeLimit      int64  `json:"fee_limit"`
}
