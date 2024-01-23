# go-xpx-check-token-util

The check token util script is a tool to detect the balance of a single asset in a single account. Alert will be sent via Telegram if asset balance is found to be less than the threshold balance.

# Getting Started

## Prerequisites

* [Golang](https://golang.org/
) is required (tested on go1.21.5)
* [Telegram Bots API](https://core.telegram.org/bots
) Key and Chat Id

## Clone the Project
```

git clone git@github.com:proximax-foundry/go-xpx-check-token-util.git
cd go-xpx-check-token-util

```

# Configurations
Configurations can be made to the script by changing the values to the fields in config.json.
```json
{
    "apiNode": "https://api-2.testnet2.xpxsirius.io/",
    "botApiKey": "1111111111:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "chatID": 111111111,
    "sleep": 60,
    "account": "VCMCJPRMJ6IUBOZ7HCYBQOSEOVGISX6AMUJ4ESTN",
    "mosaic": "13BFC518E40549D7",
    "thresholdBalance": 400000,
    "accProfile": "https://bctestnetexplorer.xpxsirius.io/#/account/"
}
```
* `apiNode`: API node URL
* `botApiKey`: Telegram Bot API Key (in String format)
* `chatID`: Telegram Chat ID (in numeric format)
* `sleep`: The time interval (in seconds) of checking the assets' balance
* `account`: List of accounts to be checked
* `mosaic`: List of assets to be checked
* `thresholdBalance`: The amount where if the asset’s balance falls lower than that, an alert will be sent
* `accProfile`: Account profile URL

Note that the default values in config.json are presented solely as examples.

# Running the Script
```go
go run main.go
```
