#!/bin/sh
MONGOHOST_FROM_HOSTNAME="$(getent hosts mongo | awk '{ print $1 }')"
MONGO_HOST="${MONGO_HOST:-$MONGOHOST_FROM_HOSTNAME}"
MONGO_PORT="${MONGO_PORT:-27017}"

CONFIG_JSON=$(cat <<EOF
{
  "Port": $PORT,
  "NetworkID": "$NETWORK_ID",
  "Mode": "$MODE",
  "Mongo": "mongodb://${MONGO_USERNAME}:${MONGO_PASSWORD}@$MONGO_HOST:$MONGO_PORT",
  "Mongodb": "$MONGO_DBNAME",
  "CoinserviceURL": "$COIN_SERVICE_URL",
  "FullnodeURL": "$FULLNODE_URL",
  "FullnodeAuthKey": "$FULLNODE_AUTHKEY",
  "ShieldService": "$SHIELD_SERVICE_URL",
  "BTCShieldPortal":"$BTC_SHIELD_URL",
  "FaucetService": "$FAUCET_SERVICE_URL",
  "CaptchaSecret":"$CAPTCHA_SECRET",
  "SlackMonitor":"$SLACK_MONITOR",
  "IncKey": "$INC_KEY",
  "EVMKey": "$EVM_KEY",
  "ISIncPrivKeys":$ISINC_Key,
  "CentralIncPaymentAddress": "$CINC_PA",
  "GGCProject": "$GOOGLE_CLOUD_PROJECT",
  "GGCAuth":"$GOOGLE_CLOUD_ACC"
}
EOF
)

printf "$CONFIG_JSON" > cfg.json

# "$@" to pass through all arguments to script
./webservice "$@"
