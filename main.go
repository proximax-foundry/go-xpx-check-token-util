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
    ApiNode             *string
    BotApiKey           *string
    ChatID              *int64
    Sleep               *int
    Account             *string
    Mosaic              *string
    ThresholdBalance    *float64
    AccProfile          *string
}

var config Config
var client *sdk.Client
var account *sdk.AccountInfo
var mosaic uint64
var divisibility float64
var mosaicName string
var currBalance float64

func main() {
    err := readConfig()
    if err != nil {
        errHandling(err)
    }
    
    conf, err := sdk.NewConfig(context.Background(), []string{*config.ApiNode})
    if err != nil {
        errHandling(fmt.Errorf("Failed to get config for HTTP Client: %v\n", err))
    }
    client = sdk.NewClient(nil, conf)

    account, mosaic, err = validateAccMosaic()
    if err != nil {
        errHandling(fmt.Errorf("Failed validation: %v\n", err))
    }

    divisibility, mosaicName = getMosaicDetails()
    for {
        for _, accMosaic := range account.Mosaics {
            mosaicIdUint64 := accMosaic.AssetId.Id()
            currBalance = float64(accMosaic.Amount)/math.Pow(10, divisibility)
            if mosaic == mosaicIdUint64 && currBalance < *config.ThresholdBalance {
                err := sendAlert()
                if err != nil {
                    errHandling(fmt.Errorf("Failed to send alert: %v\n", err))
                }
                break
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

func constructMsg() (string, []tgbotapi.MessageEntity) {
    strBalance := strconv.FormatFloat(*config.ThresholdBalance, 'f', -1, 64)
    strMsg := fmt.Sprintf(
        "Your account %s has %s balance less than %s, currently %f %s", 
        *config.Account, mosaicName, strBalance, currBalance, mosaicName)
    var entities []tgbotapi.MessageEntity
    entities = append(entities, tgbotapi.MessageEntity{
        Type: "text_link",
        Offset: 13,
        Length: 40,
        URL: *config.AccProfile + *config.Account,
    })
    return strMsg, entities
}

func errHandling(err error) {
    log.Fatal("Error: ", err)
}

func getMosaicDetails() (float64, string) {
    mosaicId, err := sdk.NewMosaicId(mosaic)
    if err != nil {
        errHandling(fmt.Errorf("Failed to get Mosaic ID: %v\n", err))
    }
    mosaicInfo, err := client.Mosaic.GetMosaicInfo(context.Background(), mosaicId)
    if err != nil {
        errHandling(fmt.Errorf("Failed to get Mosaic Info: %v\n", err))
    }
    // get divisibility
    divisibility := float64(mosaicInfo.Properties.MosaicPropertiesHeader.Divisibility)
    mosaicNames, err := client.Mosaic.GetMosaicsNames(context.Background(), mosaicId)
    if err != nil {
        errHandling(fmt.Errorf("Failed to get Mosaic Name: %v\n", err))
    }
    // get mosaic name
    mosaicName := mosaicNames[0].Names[0]
    j := strings.Index(mosaicName, ".")
    if j > -1 {
        mosaicName = mosaicName[j+1:]
    }
    return divisibility, strings.ToUpper(mosaicName)
}

func readConfig() (error) {
    configFile, err := os.Open("config.json")
    if err != nil {
        return fmt.Errorf("Cannot open config.json: %v\n", err)
    }
    defer configFile.Close()

    jsonParser := json.NewDecoder(configFile)
    if jsonParser.Decode(&config); err != nil {
        return fmt.Errorf("Cannot decode config.json: %v\n", err)
    }
    
    if err = checkMissingFields(config); err != nil {
        return err
    }
    return nil
}

func sendAlert() (error) {
    bot, err := tgbotapi.NewBotAPI(*config.BotApiKey)
    if err != nil {
        return fmt.Errorf("Failed to create BotAPI instance: %v\n", err)
    }
    bot.Debug = true
    updateConfig := tgbotapi.NewUpdate(0)
    updateConfig.Timeout = 30

    strMessage, entities := constructMsg()
    msg := tgbotapi.NewMessage(*config.ChatID, strMessage)
    msg.Entities = entities
    bot.Send(msg)
    return nil
}

func validateAccMosaic() (*sdk.AccountInfo, uint64, error) {
    address := sdk.NewAddress(*config.Account, client.NetworkType())
    account, err := client.Account.GetAccountInfo(context.Background(), address)
    if err != nil {
        return nil, 0, fmt.Errorf("Invalid account: %v\n", err)
    }

    mosaicIdUint64, err := strconv.ParseUint(*config.Mosaic, 16, 64)
    if err != nil {
        return account, mosaicIdUint64, fmt.Errorf("Failed to parse mosaic ID as Uint64: %v\n", err)
    }

    mosaicId, err := sdk.NewMosaicId(mosaicIdUint64)
    if err != nil {
        return account, mosaicIdUint64, fmt.Errorf("Failed to get Mosaic ID: %v\n", err)
    }

    _, err = client.Mosaic.GetMosaicInfo(context.Background(), mosaicId)
    if err != nil {
        return account, mosaicIdUint64, fmt.Errorf("Invalid mosaic: %v\n", err)
    }
    return account, mosaicIdUint64, nil
}