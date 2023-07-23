package provider

import (
	"github.com/shopspring/decimal"
)

type Providerer interface {
	GenerateWallet() (Wallet, error)
	GetTRXBalance(string) (decimal.Decimal, error)
	SendTRX(Wallet, string, decimal.Decimal) (string, error)
	GetTRC20Balance(string) (decimal.Decimal, error)
	SendTRC20(Wallet, string, decimal.Decimal) (string, error)
}

type Wallet struct {
	PrivKey string
	Address string
}
