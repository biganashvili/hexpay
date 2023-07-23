package main

import (
	"fmt"
	"hexpay/provider"
	"hexpay/provider/tron"
	"time"
)

const (
	fullNodeUrl         string = "https://api.trongrid.io"
	solidityNodeURL     string = "https://api.trongrid.io"
	usdtContractAddress string = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

func main() {
	var tronClient provider.Providerer
	tronClient = tron.New(fullNodeUrl, solidityNodeURL, usdtContractAddress)
	// wallet, err := usdt.GenerateWallet()
	// if err != nil {
	// 	panic(err)
	// }
	wallet := provider.Wallet{
		Address: "TFwpzzQoGTJW4hUhGKKUZe4wSVCgyMoodZ",
		PrivKey: "e8135b91771671df0b9cc9a40137660a47b9babf7539b7c55756dd6816de5f4e",
	}
	fmt.Println("new address: " + wallet.Address)

	for {
		trxBalance, err := tronClient.GetTRXBalance(wallet.Address)
		if err != nil {
			panic(err)
		}
		fmt.Println("new address TRX balance: " + trxBalance.String())
		usdtBalance, err := tronClient.GetTRC20Balance(wallet.Address)
		if err != nil {
			panic(err)
		}
		fmt.Println("new address USDT balance: " + usdtBalance.String())
		if usdtBalance.IsZero() {
			fmt.Println("SendTRC20 ERROR: zero balance")
			time.Sleep(5 * time.Second)
			continue
		}

		txID, err := tronClient.SendTRC20(wallet, "TJ3VtXGnuGJQTBqNzqA7TPtvAC999bfTAX", usdtBalance)
		if err != nil {
			fmt.Println("SendTRC20 ERROR: " + err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		fmt.Println("SendTRC20 SUCCESS: " + txID)
		break
	}

}
