

# Tonkeeper TWA API

Tonkeeper TWA API is a backend service for Tonkeeper TWA 
which is a version of [Tonkeeper](https://github.com/tonkeeper/tonkeeper-web) 
distributed as a [Telegram Web Application](https://docs.twa.dev/docs/introduction/about-platform).

# How it works

Tonkeeper TWA API aggregates data from two sources:
* from [HTTP Bridge](https://github.com/ton-blockchain/ton-connect/blob/main/bridge.md) which is a transport mechanism to deliver messages from an app to a wallet and vice versa.
* from [TonAPI.io](https://tonapi.io/) which is a service that notifies about new events in the TON blockchain and provides a convenient way to interact with the blockchain.

Tonkeeper TWA API exposes a [REST API](api/tonkeeper-twa-api.yml) 
to authenticate telegram users and their wallets and subscribe to events from the Bridge and TonAPI.io.

![How it works](docs/twa-api.png)

# Installation

Environment variables:

| Env variable              | Description                                                                                                                                                                        |
|---------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `PORT`                    | Port to listen on, the default is `7077`                                                                                                                                           |
| `TON_API_KEY`             | An optional tonapi.io API key. You can get it at [TonConsole](https://tonconsole.com/). Without a key, the TonAPI will work in a limited mode.                                     |
| `TON_CONNECT_SECRET`      | A secret key that is unique per installation. Used in authentication process to verify ownership of a wallet.                                                                      |
| `TELEGRAM_BOT_SECRET_KEY` | A secret key of your telegram bot. Used to work with telegram API and process [twa init data](https://docs.twa.dev/docs/launch-params/init-data#authorization-and-authentication). |
| `POSTGRES_URI`            | A connection string for a postgres database. An example: `postgres://user:password@localhost:5432/tonkeeper_twa_api`                                                               |


TODO: how to run it in docker