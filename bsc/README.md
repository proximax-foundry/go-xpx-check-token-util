
To check the token balance of a Binance Smart Chain (BSC) account using Go (Golang), you can use the go-ethereum library, which provides tools for interacting with Ethereum-based blockchains, including BSC. Below is an example script that you can use as a starting point. Make sure to install the go-ethereum library if you haven't already:

```bash
go get -u github.com/ethereum/go-ethereum
```

Replace placeholders like 0xYourBSCAddressHere and 0xTokenContractAddressHere with the actual BSC address and token contract address you want to check. The script uses the eth_getBalance RPC call for BNB balance and the eth_call RPC call for token balance.

Note: The example assumes that the token you are querying has 18 decimal places. If the token has a different number of decimal places, you may need to adjust the getTokenBalance function accordingly. Additionally, error handling and other considerations may need to be addressed based on the specific requirements of your application.