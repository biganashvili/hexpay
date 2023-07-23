package tron

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hexpay/provider"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

const (
	BalanceFuncSelector string = "70a08231"
)

type trc20 struct {
	FullNodeUrl       string
	SolidityNodeURL   string
	ContractAddress   string
	Denomination      decimal.Decimal
	TRC20Denomination decimal.Decimal
}

func New(fullNodeUrl, solidityNodeURL, contractAddress string) *trc20 {
	return &trc20{
		FullNodeUrl:       fullNodeUrl,
		SolidityNodeURL:   solidityNodeURL,
		ContractAddress:   contractAddress,
		Denomination:      decimal.NewFromInt(1000000),
		TRC20Denomination: decimal.NewFromInt(1000000),
	}
}

func (client trc20) GenerateWallet() (provider.Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return provider.Wallet{}, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	hexPrivateKey := hexutil.Encode(privateKeyBytes)[2:]
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return provider.Wallet{}, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	address = "41" + address[2:]
	addb, err := hex.DecodeString(address)
	if err != nil {
		return provider.Wallet{}, err
	}
	hash1 := s256(s256(addb))
	secret := hash1[:4]
	addb = append(addb, secret...)
	base58Address := base58.Encode(addb)

	if err != nil {
		return provider.Wallet{}, err
	}

	return provider.Wallet{PrivKey: hexPrivateKey, Address: base58Address}, nil
}

func (client *trc20) SendTRC20(wallet provider.Wallet, toAddress string, amount decimal.Decimal) (string, error) {

	value := fmt.Sprintf("%x", amount.Mul(client.TRC20Denomination).IntPart())
	payload := `
	{
		"owner_address": "` + client.Base58ToHex(wallet.Address) + `",
		"contract_address": "` + client.Base58ToHex(client.ContractAddress) + `", 
		"function_selector":"transfer(address,uint256)",
		"parameter":"` + strings.Repeat("0", 24) + client.Base58ToHex(toAddress)[2:] + strings.Repeat("0", 64-len(value)) + value + `",
		"call_value":0,
		"fee_limit":10000000000
	}`

	response, status, err := jsonRPC([]byte(payload), client.SolidityNodeURL+"/wallet/triggersmartcontract", "POST")
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", fmt.Errorf("SendTRC20 Code: %d, Body: %s", status, string(response))
	}
	var result struct {
		Transaction RawTransaction
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return "", err
	}

	if result.Transaction.TxID == "" {
		return "", fmt.Errorf("SendTRC20 response: %s", string(response))
	}

	pk, err := crypto.HexToECDSA(wallet.PrivKey)
	if err != nil {
		return "", err
	}
	result.Transaction.Visible = false
	signedTransaction, err := client.signRawTransaction(&result.Transaction, pk)
	if err != nil {
		return "", err
	}

	return client.broadcastTransaction(signedTransaction)
}

func jsonRPC(data []byte, url string, requestType string) ([]byte, int, error) {
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(data))
	if err != nil {
		return []byte{}, 0, nil
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, err
	}

	return body, resp.StatusCode, nil
}

func (client *trc20) GetTRC20Balance(address string) (decimal.Decimal, error) {
	balance := decimal.Zero
	data := `
	{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "eth_call",
		"params": [
			{
				"to":    "0x` + client.Base58ToHex(client.ContractAddress)[2:] + `",
				"value": "0x0",
				"data":  "0x` + BalanceFuncSelector + strings.Repeat("0", 24) + client.Base58ToHex(address)[2:] + `"
			},
			"latest"
		]
	}`

	response, status, err := jsonRPC(
		[]byte(data),
		client.SolidityNodeURL+"/jsonrpc",
		"POST",
	)

	if err != nil {
		return balance, err
	}
	if status != 200 {
		return balance, fmt.Errorf("GetTRC20Balance status: %d,  body: %s", status, string(response))
	}

	var result struct {
		ID      any    `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
		Result string `json:"result,omitempty"`
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return balance, err
	}
	if result.Result == "" {
		return balance, fmt.Errorf("GetTRC20Balance body: %s", string(response))
	}

	i := new(big.Int)
	i.SetString(result.Result[2:], 16)
	sun := decimal.NewFromBigInt(i, 0)
	balance = sun.DivRound(client.Denomination, 18)
	return balance, nil

}

func (client *trc20) SendTRX(wallet provider.Wallet, to string, amount decimal.Decimal) (string, error) {
	var rawTransaction RawTransaction
	amount = amount.Mul(client.Denomination)
	data := []byte(
		`{
			"owner_address": "` + wallet.Address + `",
			"to_address": "` + to + `",
			"amount": ` + fmt.Sprintf("%d", amount.IntPart()) + `,
			"visible": true
		  }`,
	)
	response, status, err := jsonRPC(
		data,
		client.FullNodeUrl+"/wallet/createtransaction",
		"POST",
	)
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", fmt.Errorf("createTransaction status: %d,  body: %s", status, string(response))
	}
	err = json.Unmarshal(response, &rawTransaction)
	if err != nil {
		return "", err
	}
	if rawTransaction.TxID == "" {
		return "", fmt.Errorf(string(response))
	}

	pk, err := crypto.HexToECDSA(wallet.PrivKey)
	if err != nil {
		return "", err
	}

	signedTransaction, err := client.signRawTransaction(&rawTransaction, pk)
	if err != nil {
		return "", err
	}

	return client.broadcastTransaction(signedTransaction)
}

func (client *trc20) signRawTransaction(tx *RawTransaction, key *ecdsa.PrivateKey) (*SignedTransaction, error) {
	rawData, err := json.Marshal(tx.RawData)
	if err != nil {
		return &SignedTransaction{}, err
	}

	signedTransaction := &SignedTransaction{
		Visible:    tx.Visible,
		TxID:       tx.TxID,
		RawData:    string(rawData),
		RawDataHex: tx.RawDataHex,
	}
	txIDbytes, err := hex.DecodeString(tx.TxID)
	if err != nil {
		return &SignedTransaction{}, err
	}

	signature, err := crypto.Sign(txIDbytes, key)
	if err != nil {
		return &SignedTransaction{}, err
	}

	signedTransaction.Signature = append(signedTransaction.Signature, hex.EncodeToString(signature))
	return signedTransaction, nil
}

func (client *trc20) broadcastTransaction(tx *SignedTransaction) (string, error) {
	js, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	response, status, err := jsonRPC(js, client.SolidityNodeURL+"/wallet/broadcasttransaction", "POST")
	if err != nil {
		return "", err
	}
	if status != 200 {
		return "", fmt.Errorf("broadcastTransaction status: %d,  body: %s", status, string(response))
	}

	var result struct {
		Result bool
		Txid   string
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return "", err
	}
	if !result.Result || result.Txid == "" {
		return "", fmt.Errorf("broadcastTransaction response: %s", string(response))
	}

	return result.Txid, nil
}

func (client *trc20) GetTRXBalance(address string) (decimal.Decimal, error) {

	data := `{
			"id":      1,
			"jsonrpc": "2.0",
			"method":  "eth_getBalance",
			"params":  ["0x` + client.Base58ToHex(address) + `", "latest"]
	}`

	response, status, err := jsonRPC(
		[]byte(data),
		client.SolidityNodeURL+"/jsonrpc",
		"POST",
	)

	if err != nil {
		return decimal.Zero, err
	}
	if status != 200 {
		return decimal.Zero, fmt.Errorf("GetTRXBalance status: %d,  body: %s", status, string(response))
	}

	type Result struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      any    `json:"id"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
		Result string `json:"result,omitempty"`
	}
	result := Result{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return decimal.Zero, err
	}
	if result.Error.Code != 0 || result.Error.Message != "" {
		return decimal.Zero, fmt.Errorf("%s", result.Error.Message)
	}
	if len(result.Result) < 3 {
		return decimal.Zero, fmt.Errorf("%s", "unexpected response")
	}

	if len(result.Result) <= 2 {
		return decimal.Zero, errors.New("balance value not found")
	}
	i := new(big.Int)
	i.SetString(result.Result[2:], 16)
	sun := decimal.NewFromBigInt(i, 0)
	trx := sun.DivRound(client.Denomination, 18)
	return trx, nil
}

func s256(s []byte) []byte {
	h := sha256.New()
	h.Write(s)
	bs := h.Sum(nil)
	return bs
}

func (client trc20) Base58ToHex(address string) string {
	//convert base58 to hex
	decodedAddress := base58.Decode(address)
	dst := make([]byte, hex.EncodedLen(len(decodedAddress)))
	hex.Encode(dst, decodedAddress)
	dst = dst[:len(dst)-8]
	return string(dst)
}
