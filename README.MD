# Vega BOT

Community project aiming to simplify the creating of Vega bots.
This framework contains common helpers to create a bot and it's configurable in order to make at simple as possible to create your bot.

The bot is retrieving candles and executing a custom script each time a new candle starts. It runs over a single market and it can store candle backlogs in order to perform

## Configuration
The bot is configurable by editing the config.yaml file. Here is a sample content:

```yaml
# Vega parameters
GrpcNodeUrl: "n06.testnet.vega.xyz:3002"            # The Vega gRPC node url
VegaEventsBatchSize: 15000                          # Vega client gRPC configuration. For more info check the doc: https://docs.fairground.vega.xyz/docs/api-howtos/event-stream/
Debug: true                                         # True if you want to enable verbose logging
MarketName: "Tesla Quarterly (31 Dec 2021)"         # The name of Vega market 

# Wallet parameters
WalletServerURL: "http://localhost:1789"            # Vega wallet url. Set to https://wallet.testnet.vega.xyz if you want to use Vega hosted wallet.
WalletName: "user"                                  # Vega wallet username
WalletPassphrase: "aa"                              # Vega wallet password

# Sentry parameters
SentryEnabled: true                                 # Configure to true if you want to log crash stracktraces to Sentry (advanged feature)
SentryDsn: ""                                       # The Sentry endpoint for the project.

# Prometheus parametes
PrometheusEnabled: true                             # Enable it if you want the bot to expose a Prometheus endpoint
PrometheusPort: 2112                                # Prometheus endpoint port

# Strategy config
CandlesBacklog: 5                                   # The number of candles to keep as backlog
Strategy: "debug"                                   # The stategy to execute. The strategy function should have the same name and be stored in the root folder of the project.
TransactionQuantity: 2                              # The number of assets to buy in each transaction. This value can be dynamically overwritten at runtine.
```

## Create your own strategy
In order to create a new file in the project's root folder and create the strategy main function. This function must accept a parameter containing the list of candles and must return:

- Long: a boolean value set to true in case the bot would send a long signal;
- Short: a boolean value set to true in case the bot would send a short signal;
- Price: an unsigned int64 containing the transaction price per unit;
- Error: an error variable containing an error in case the execution doesn't succed;

For example a strategy entrypoint should look as the following:

```go
func myStrategy(cache list.List) (bool, bool, uint64, error) 
```

## TODO

- Add backtesting feature;
- Add social notifications;
- Start/Stop/Configure using Telegram bot;


# About Vega
[Vega](https://vega.xyz) is a protocol for creating and trading derivatives on a fully decentralised network. The network, secured with proof-of-stake, will facilitate fully automated, end-to-end margin trading and execution of complex financial products. Anyone will be able to build decentralised markets using the protocol.

Read more at [https://vega.xyz](https://vega.xyz).