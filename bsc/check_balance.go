package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	// Replace with your BSC RPC endpoint
	rpcEndpoint := "https://bsc-dataseed.binance.org/"

	// Replace with the BSC address you want to check
	address := "0xYourBSCAddressHere"

	// Replace with the token contract address you want to check
	tokenContractAddress := "0xTokenContractAddressHere"

	// Connect to the BSC node
	client, err := rpc.Dial(rpcEndpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Get the balance of BNB (native coin on BSC)
	balance, err := getBNBBalance(client, address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("BNB Balance for %s: %s BNB\n", address, balance)

	// Get the balance of a specific token
	tokenBalance, err := getTokenBalance(client, tokenContractAddress, address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Token Balance for %s: %s Tokens\n", address, tokenBalance)
}

func getBNBBalance(client *rpc.Client, address string) (*big.Float, error) {
	var result hexOrDecimalBigInt
	err := client.CallContext(context.Background(), &result, "eth_getBalance", address, "latest")
	if err != nil {
		return nil, err
	}

	balanceWei, success := new(big.Int).SetString(result.String(), 0)
	if !success {
		return nil, fmt.Errorf("failed to convert balance to big.Int")
	}

	// Convert from Wei to BNB
	balanceBNB := new(big.Float).Quo(new(big.Float).SetInt(balanceWei), big.NewFloat(1e18))
	return balanceBNB, nil
}

func getTokenBalance(client *rpc.Client, tokenContractAddress string, address string) (*big.Float, error) {
	// Create the data field for the balanceOf function of ERC20 tokens
	data := fmt.Sprintf("0x70a08231000000000000000000000000%s", address[2:]) // Remove "0x" prefix

	var result hexOrDecimalBigInt
	err := client.CallContext(context.Background(), &result, "eth_call", ethereum.CallMsg{
		To:   &tokenContractAddress,
		Data: []byte(data),
	}, "latest")
	if err != nil {
		return nil, err
	}

	balance, success := new(big.Int).SetString(result.String(), 0)
	if !success {
		return nil, fmt.Errorf("failed to convert balance to big.Int")
	}

	// Assuming the token has 18 decimal places, you may need to adjust this based on the specific token
	balanceFloat := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
	return balanceFloat, nil
}

type hexOrDecimalBigInt struct {
	*big.Int
}

func (h *hexOrDecimalBigInt) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str[:2] == "0x" {
		// Hex
		i, success := new(big.Int).SetString(str[2:], 16)
		if !success {
			return fmt.Errorf("failed to parse hex number")
		}
		h.Int = i
	} else {
		// Decimal
		i, success := new(big.Int).SetString(str, 10)
		if !success {
			return fmt.Errorf("failed to parse decimal number")
		}
		h.Int = i
	}
	return nil
}
