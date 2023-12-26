package main

import (
    "math"
    "time"
    "context"
    "os"
    "encoding/json"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/proximax-storage/go-xpx-chain-sdk/sdk"
)

type Config struct {
    ApiNode       string   `json:"apiNode"`
    Sleep         int      `json:"sleep"`
    ChatID        int64    `json:"chatID"`
}

var latestTxnHash string

func main() {
    config := readConfig("config.json")
    conf, err := sdk.NewConfig(context.Background(),[]string{config.ApiNode})
    errHandling(err)
    client := sdk.NewClient(nil, conf)

    for {
        txns, err := client.Transaction.GetTransactionsByGroup(context.Background(), sdk.Confirmed,     
            &sdk.TransactionsPageOptions{
                PaginationOrderingOptions: sdk.PaginationOrderingOptions{
                    SortDirection: "desc",
                },
            })
        errHandling(err)

        newTxn := txns.Transactions[0]
        newTxnHash := newTxn.GetAbstractTransaction().TransactionInfo.TransactionHash.String()

        if newTxnHash != latestTxnHash {
            accsAddress := []*sdk.Address{}
            
            for _, txn := range txns.Transactions {
                hash := txn.GetAbstractTransaction().TransactionInfo.TransactionHash.String()
                if hash != latestTxnHash {
                    currAddress := txn.GetAbstractTransaction().Signer.Address.Address
                    multisigAccInfo, err := client.Account.GetMultisigAccountInfo(context.Background(), &sdk.Address{
                        Address: currAddress,})
                    if (err != nil) {
                        continue
                    }
                    if len(multisigAccInfo.MultisigAccounts) > 0 {
                        accsAddress = append(accsAddress, &sdk.Address{
                            Address: currAddress,
                        })
                    }
                } else {
                    break
                }
            }
            latestTxnHash = newTxnHash
            if len(accsAddress) > 0 {
                accsInfo, err := client.Account.GetAccountsInfo(context.Background(), accsAddress...)
                errHandling(err)
    
                targetAccs := []*string{}
                for _, acc := range accsInfo {
                    if float64(acc.Mosaics[0].Amount)/math.Pow(10, 6) < 500 {
                        targetAccs = append(targetAccs, &acc.Address.Address)
                    }
                }
                if len(targetAccs) > 0 {
                    sendAlert(targetAccs...)
                }
            }
        }
        time.Sleep(time.Duration(config.Sleep) * time.Second)
    }
}

func errHandling(err error) {
    if err != nil {
        panic(err)
    }
}

func readConfig(fileName string) (Config) {
    configFile, err := os.Open(fileName)
    errHandling(err)

    var config Config
    defer configFile.Close()

    jsonParser := json.NewDecoder(configFile)
    err = jsonParser.Decode(&config)
    errHandling(err)

    return config
}

func sendAlert(targets ...*string) {
    config := readConfig("config.json")

    bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
    errHandling(err)
    bot.Debug = true
    updateConfig := tgbotapi.NewUpdate(0)
    updateConfig.Timeout = 30

    strMessage := ""
    entities := []tgbotapi.MessageEntity{}
    for i, target := range targets {
        strMessage += *target + "\n"

        entities = append(entities, tgbotapi.MessageEntity{
            Type: "text_link",
            Offset: i*41,
            Length: 40,
            URL: "https://bctestnetexplorer.xpxsirius.io/#/account/" + *target,
        })
    }
    strMessage += "\nYour wallet currently has less than 500 XPX. Do top up to avoid inconvenience in making transactions."

    msg := tgbotapi.NewMessage(config.ChatID, strMessage)
    msg.Entities = entities
    bot.Send(msg)
}