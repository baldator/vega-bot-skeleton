package main

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"code.vegaprotocol.io/go-wallet/wallet"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto"
	"github.com/vegaprotocol/api-clients/go/generated/code.vegaprotocol.io/vega/proto/api"
	vegagrpc "github.com/vegaprotocol/api/grpc/clients/go/grpc"
	"google.golang.org/grpc"
)

type strategyMapping map[string]interface{}
type StrategyResult struct {
	Short bool   `json:"short"`
	Long  bool   `json:"long"`
	Price uint64 `json:"price"`
}

var strategies = strategyMapping{}

func main() {
	// Read application config
	conf, err := ReadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}

	if conf.SentryEnabled {
		initializeSentry(conf.SentryDsn)
	}

	if conf.PrometheusEnabled {
		initializePrometheus(conf.PrometheusPort)
	}

	// setup wallet connection
	walletserverURL := CheckWalletUrl(conf.WalletServerURL)
	walletConfig := WalletConfig{
		URL:        walletserverURL,
		Name:       conf.WalletName,
		Passphrase: conf.WalletPassphrase,
	}

	var token wallet.TokenResponse
	body, err := LoginWallet(walletConfig)
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(body), &token)

	fmt.Println(token.Token)

	if err != nil {
		panic(err)
	}

	var pubkey wallet.Keypair
	keysResp, err := GetKeyPairs(walletConfig, token.Token)
	if err != nil {
		panic(err)
	}
	var keys wallet.KeysResponse
	json.Unmarshal([]byte(keysResp), &keys)
	if conf.WalletPubKey == "" {

		pubkey = keys.Keys[0]
	} else {
		for _, key := range keys.Keys {
			if key.Pub == conf.WalletPubKey {
				pubkey = key
				break
			}
		}
	}

	// setup gRPC connection
	nodeURLGrpc := conf.GrpcNodeURL
	if len(nodeURLGrpc) == 0 {
		panic("NODE_URL_GRPC is null or empty")
	}

	conn, err := grpc.Dial(nodeURLGrpc, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataClient := api.NewTradingDataServiceClient(conn)
	tradingClient := api.NewTradingServiceClient(conn)

	// Initialize candle cache
	cache := list.New()

	// Request the identifier for a market
	request := api.MarketsRequest{}
	markets, err := dataClient.Markets(context.Background(), &request)
	if err != nil {
		panic(err)
	}
	marketID, err := getMarketID(markets.Markets, conf.MarketName)

	if err != nil {
		panic(err)
	}

	// Load strategies
	strategies = map[string]interface{}{
		"debug":     debug,
		"strategy2": strategy2,
	}

	fmt.Printf("Found market: %s \n", marketID)

	fmt.Println("Connecting to stream...")

	candleRequest := api.CandlesSubscribeRequest{MarketId: marketID, Interval: proto.Interval_INTERVAL_I1M}
	event, err := dataClient.CandlesSubscribe(context.Background(), &candleRequest)

	done := make(chan bool)

	fmt.Printf("Start execution:\n")
	fmt.Printf("  - Strategy: %s\n", conf.Strategy)
	fmt.Printf("  - Candle backlog: %d\n", conf.CandlesBacklog)
	fmt.Printf("  - Market name: %s\n", conf.MarketName)

	go func() {
		var prevCandle *proto.Candle
		fmt.Println("Listening to events...")
		for {
			resp, err := event.Recv()

			if err == io.EOF {
				// read done.
				close(done)
				return
			}
			if err != nil {
				log.Printf("%+v\n", err)
				log.Printf("%+v\n", vegagrpc.ErrorDetail(err))
				panic(err)
			}
			//fmt.Printf("Resp received: %v\n", resp)

			if prevCandle != nil && resp.Candle.Timestamp != prevCandle.Timestamp {
				fmt.Printf("New candle\n")
				// add element to queue and remove last element if needed
				cache.PushBack(prevCandle) // Enqueue

				for cache.Len() > conf.CandlesBacklog {
					e := cache.Front() // First element
					cache.Remove(e)    // Dequeue
				}

				// call strategy
				result, _ := Call(conf.Strategy, *cache)

				fmt.Printf("Strategy result: %+v\n", result)
				fmt.Printf("  - Long: %t\n", result.Long)
				fmt.Printf("  - Short: %t\n", result.Short)

				if result.Long {
					err = submitOrder(marketID, 1, result.Price, proto.Side_SIDE_BUY, walletConfig, token.Token, pubkey.Pub, dataClient, tradingClient)
					if err != nil {
						log.Printf("%+v\n", err)
						log.Printf("%+v\n", vegagrpc.ErrorDetail(err))
						panic(err)
					}
				}

				if result.Short {
					err = submitOrder(marketID, 1, result.Price, proto.Side_SIDE_SELL, walletConfig, token.Token, pubkey.Pub, dataClient, tradingClient)
					if err != nil {
						log.Printf("%+v\n", err)
						log.Printf("%+v\n", vegagrpc.ErrorDetail(err))
						panic(err)
					}
				}
			}
			prevCandle = resp.Candle
		}
	}()

	defer event.CloseSend()

	<-done //we will wait until all response is received
	// :stream_events__
	fmt.Printf("finished")

}
