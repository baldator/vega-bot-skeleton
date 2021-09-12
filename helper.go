package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto/api"
)

func initializeSentry(sentryDsn string) {
	log.Println("Initialize sentry")
	if sentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: sentryDsn,
		})
		if err != nil {
			log.Fatalf("sentry.Init: %s", err)
		}
	}
	defer sentry.Flush(2 * time.Second)
}

func initializePrometheus(port int) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("listen on %d\n", port)
		portString := ":" + strconv.Itoa(port)
		log.Fatal(http.ListenAndServe(portString, nil))
	}()
}

func getMarketID(markets []*proto.Market, marketName string) (string, error) {
	for _, m := range markets {
		log.Printf("Market: %s, target: %s\n", m.TradableInstrument.Instrument.Name, marketName)

		if m.TradableInstrument.Instrument.Name == marketName {
			return m.GetId(), nil
		}
	}

	return "", fmt.Errorf("Market name cannot be found")
}

func Call(funcName string, params ...interface{}) (result *StrategyResult, err error) {
	f := reflect.ValueOf(strategies[funcName])
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is out of index.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	var res []reflect.Value
	res = f.Call(in)
	result = &StrategyResult{
		Short: res[0].Bool(),
		Long:  res[1].Bool(),
		Price: res[2].Uint(),
	}
	return
}

func submitOrder(marketID string, quantity uint64, price uint64, side proto.Side, walletConfig WalletConfig, walletToken string, walletPubkey string, dataClient api.TradingDataServiceClient, tradingClient api.TradingServiceClient) error {
	request := api.GetVegaTimeRequest{}
	vegaTime, err := dataClient.GetVegaTime(context.Background(), &request)

	if err != nil {
		return err
	}

	expireAt := vegaTime.Timestamp + (120 * 1e9)
	// :get_expiry_time__
	fmt.Printf("Blockchain time: %d\n", vegaTime.Timestamp)
	fmt.Printf("Order expiration time: %d\n", expireAt)

	// Submit order
	// __prepare_submit_order:
	// Prepare a submit order message
	orderSubmission := proto.OrderSubmission{
		Size:        quantity,
		Price:       price,
		MarketId:    marketID,
		Side:        side,
		TimeInForce: proto.Order_TIME_IN_FORCE_GTT,
		Type:        proto.Order_TYPE_MARKET,
	}

	order := api.PrepareSubmitOrderRequest{Submission: &orderSubmission}
	orderRequest, err := tradingClient.PrepareSubmitOrder(context.Background(), &order)
	if err != nil {
		return err
	}
	data := orderRequest.Blob
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	_, err = SignTransaction(walletConfig, walletToken, walletPubkey, string(sEnc))
	if err != nil {
		panic(err)
	}

	return nil
}
