# go-xpx-check-token-util

The check token util script is a tool to detect the balance of assets in specific accounts. Alert will be sent to the relevant accounts via Telegram to top up their assets if they are found to be less than the specified amount.

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
    "botApiKey": "<TELEGRAM_BOT_API_KEY>",
    "chatID": <TELEGRAM_CHAT_ID>,
    "minAmount": 500,
    "sleep": 60,
    "accounts": [
        "VCMCJPRMJ6IUBOZ7HCYBQOSEOVGISX6AMUJ4ESTN",
        "VCYV4I-MG7FEN-QNRAAR-MEHAZE-ND5AI2-V4325W-RNRM"
    ],
    "mosaics": [
        "13BFC518E40549D7",
        "705BAFA9B6903C08",
        "2D829694552B1189",
        "52FC262C2AB5CAE5",
        "69DD64ED4343011C"
    ]
}
```
* `apiNode`: API node URL
* `botApiKey`: Telegram Bot API Key (in String format)
* `chatID`: Telegram Chat ID (in numeric format)
* `minAmount`: The amount where if the asset’s balance falls lower than that, an alert will be sent
* `sleep`: The time interval (in seconds) of checking the assets' balance
* `accounts`: List of accounts to be checked
* `mosaics`: List of mosaics to be checked

# Running the Script
```go
go run main.go
```
