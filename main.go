package main

import (
    "os"
    "fmt"
    "log"
    "math"
    "time"
    "context"
    "strconv"
    "strings"
    "reflect"
    "encoding/json"
    "github.com/proximax-storage/go-xpx-chain-sdk/sdk"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
    ApiNode       *string
    BotApiKey     *string
    ChatID        *int64
    MinAmount     *float64
    Sleep         *int
    Accounts      []*string
    Mosaics       []*string
}

type MosaicDetail struct {
    Account       string
    Names         []*string
}

func main() {
    config, err := readConfig()
    if err != nil {
        errHandling(err)
    }
    
    conf, err := sdk.NewConfig(context.Background(), []string{*config.ApiNode})
    if err != nil {
        errHandling(fmt.Errorf("Failed to get config for HTTP Client: %v\n", err))
    }
    client := sdk.NewClient(nil, conf)

    accountsToCheck, mosaicsToCheck, err := validateAccMosaic(config, client)
    if err != nil {
        errHandling(fmt.Errorf("Failed validation: %v\n", err))
    }

    accsInfo, err := client.Account.GetAccountsInfo(context.Background(), accountsToCheck...)
    if err != nil {
        errHandling(fmt.Errorf("Failed to get Accounts Info: %v\n", err))
    }

    for {
        targetAccs := []*MosaicDetail{}
        for _, accInfo := range accsInfo {
            accMosaics := accInfo.Mosaics
            var mosaicNames []*string
            for _, accMosaic := range accMosaics {
                mosaicIdUint64 := accMosaic.AssetId.Id()
                for i, _ := range mosaicsToCheck {
                    if mosaicsToCheck[i] == mosaicIdUint64 {
                        mosaicId, err := sdk.NewMosaicId(mosaicIdUint64)
                        if err != nil {
                            errHandling(fmt.Errorf("Failed to get Mosaic ID: %v\n", err))
                        }
                        mosaicInfo, err := client.Mosaic.GetMosaicInfo(context.Background(), mosaicId)
                        if err != nil {
                            errHandling(fmt.Errorf("Failed to get Mosaic Info: %v\n", err))
                        }
                        divisibility := float64(mosaicInfo.Properties.MosaicPropertiesHeader.Divisibility)
                        if float64(accMosaic.Amount)/math.Pow(10, divisibility) < *config.MinAmount {
                            mosaic, err := client.Mosaic.GetMosaicsNames(context.Background(), mosaicId)
                            if err != nil {
                                errHandling(fmt.Errorf("Failed to get Mosaic Name: %v\n", err))
                            }
                            mosaicName := mosaic[0].Names[0]
                            j := strings.Index(mosaicName, ".")
                            if j > -1 {
                                mosaicName = mosaicName[j+1:]
                            }
                            mosaicNames = append(mosaicNames, &mosaicName)
                        }
                    } 
                }
            }
            if len(mosaicNames) > 0 {
                targetAccs = append(targetAccs, &MosaicDetail{
                    Account: accInfo.Address.Address,
                    Names: mosaicNames,
                })
            }
        }
        if len(targetAccs) > 0 {
            err := sendAlert(config, targetAccs...)
            if err != nil {
                errHandling(fmt.Errorf("Failed to send alert: %v\n", err))
            }
        }
        time.Sleep(time.Duration(*config.Sleep) * time.Second)
    }
}

func checkMissingFields(config Config) (error) {
    var missingFields []string

    r := reflect.ValueOf(&config).Elem()
	rt := r.Type()
    for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		rv := reflect.ValueOf(&config)
		value := reflect.Indirect(rv).FieldByName(field.Name)
        if value.IsNil() {
            missingFields = append(missingFields, field.Name)
        }
	}
    if len(missingFields) > 0 {
        errMsg := "Cannot get the value of "
        for j, field := range missingFields {
            errMsg += field
            if j != len(missingFields)-1 {
                errMsg += ", "
            }
        }
        errMsg += " in config.json."
        return fmt.Errorf(errMsg)
    }
    return nil
}

func constructMsg(config Config, targets ...*MosaicDetail) (string, []tgbotapi.MessageEntity, error) {
    n := 0
    strMsg := "The following accounts' assets have balance less than " + strconv.FormatFloat(*config.MinAmount, 'f', -1, 64) + ": \n\n"
    strMsgLen := len(strMsg)
    entities := []tgbotapi.MessageEntity{}
    for i, target := range targets {
        entities = append(entities, tgbotapi.MessageEntity{
            Type: "text_link",
            Offset: i*41 + n + strMsgLen,
            Length: 40,
            URL: "https://bctestnetexplorer.xpxsirius.io/#/account/" + target.Account,
        })

        strMsg += target.Account + "\n"
        for _, mosaicName := range target.Names {
            strMsg += "- " + strings.ToUpper(*mosaicName) + "\n"
            n += len(*mosaicName) + 3
        }
    }
    return strMsg, entities, nil
}

func errHandling(err error) {
    log.Fatal("Error: ", err)
}

func readConfig() (Config, error) {
    configFile, err := os.Open("config.json")

    if err != nil {
        return Config{}, fmt.Errorf("Cannot open config.json: %v\n", err)
    }

    var config Config
    defer configFile.Close()

    jsonParser := json.NewDecoder(configFile)
    if jsonParser.Decode(&config); err != nil {
        return Config{}, fmt.Errorf("Cannot decode config.json: %v\n", err)
    }
    
    if err = checkMissingFields(config); err != nil {
        return config, err
    }
    return config, nil
}

func sendAlert(config Config, targets ...*MosaicDetail) (error) {
    bot, err := tgbotapi.NewBotAPI(*config.BotApiKey)
    if err != nil {
        return fmt.Errorf("Failed to create BotAPI instance: %v\n", err)
    }
    bot.Debug = true
    updateConfig := tgbotapi.NewUpdate(0)
    updateConfig.Timeout = 30

    strMessage, entities, err := constructMsg(config, targets...)
    if err != nil {
        return fmt.Errorf("Failed to construct Telegram message: %v\n", err)
    }
    msg := tgbotapi.NewMessage(*config.ChatID, strMessage)
    msg.Entities = entities
    bot.Send(msg)
    return nil
}

func validateAccMosaic(config Config, client *sdk.Client) ([]*sdk.Address, []uint64, error) {
    var accounts []*sdk.Address
    var mosaics []uint64

    for _, acc := range config.Accounts {
        address := sdk.NewAddress(*acc, client.NetworkType())
        _, err := client.Account.GetAccountInfo(context.Background(), address)
        if err != nil {
            return accounts, mosaics, fmt.Errorf("Invalid account: %v\n", err)
        }
        accounts = append(accounts, address)
    }

    for _, mosaic := range config.Mosaics {
        mosaicIdUint64, err := strconv.ParseUint(*mosaic, 16, 64)
        if err != nil {
            return accounts, mosaics, fmt.Errorf("Failed to parse mosaic ID as Uint64: %v\n", err)
        }
        mosaicId, err := sdk.NewMosaicId(mosaicIdUint64)
        if err != nil {
            return accounts, mosaics, fmt.Errorf("Failed to get Mosaic ID: %v\n", err)
        }
        _, err = client.Mosaic.GetMosaicInfo(context.Background(), mosaicId)
        if err != nil {
            return accounts, mosaics, fmt.Errorf("Invalid mosaic: %v\n", err)
        }
        mosaics = append(mosaics, mosaicIdUint64)
    }
    return accounts, mosaics, nil
}