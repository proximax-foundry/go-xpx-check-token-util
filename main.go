package main

import (
	"fmt"
	"context"
	"github.com/proximax-storage/go-xpx-chain-sdk/sdk"
)

func main() {
	conf, err := sdk.NewConfig(context.Background(),[]string{"https://api-2.testnet2.xpxsirius.io/"})
	errHandling(err)
	client := sdk.NewClient(nil, conf)
	transactionPage, err := client.Transaction.GetTransactionsByGroup(
		context.Background(),
		sdk.Confirmed,
		&sdk.TransactionsPageOptions{
			// to get AggregateBonded txns
			Type: []uint{16961},
		},
	)
	errHandling(err)
	fmt.Println(transactionPage)
}

func errHandling(err error) {
    if err != nil {
        panic(err)
    }
}