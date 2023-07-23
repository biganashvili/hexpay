# hexpay
A simple example of how to create->sign->broadcast a TRC20 token


## Implemented Functionality
```
type Providerer interface {
	GenerateWallet() (Wallet, error)
	GetTRXBalance(string) (decimal.Decimal, error)
	SendTRX(Wallet, string, decimal.Decimal) (string, error)
	GetTRC20Balance(string) (decimal.Decimal, error)
	SendTRC20(Wallet, string, decimal.Decimal) (string, error)
}
```