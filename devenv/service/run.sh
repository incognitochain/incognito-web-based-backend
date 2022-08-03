#!/bin/sh
MONGOHOST_FROM_HOSTNAME="$(getent hosts mongo | awk '{ print $1 }')"
MONGO_HOST="${MONGO_HOST:-$MONGOHOST_FROM_HOSTNAME}"
MONGO_PORT="${MONGO_PORT:-27017}"

CONFIG_JSON=$(cat <<EOF
{
  "Port": $PORT,
  "NetworkID": "$NETWORK_ID",
  "Mode": "$MODE",
  "Mongo": "mongodb://${MONGO_USERNAME}:${MONGO_PASSWORD}@$MONGOHOST:$MONGO_PORT",
  "Mongodb": "$MONGO_DBNAME",
  "CoinserviceURL": "$COIN_SERVICE_URL",
  "FullnodeURL": "$FULLNODE_URL",
  "ShieldService": "$SHIELD_SERVICE_URL",
  "FaucetService": "$FAUCET_SERVICE_URL",
  "CaptchaSecret":"$CAPTCHA_SECRET",
  "IncKey": "$INC_KEY",
  "EVMKey": "$EVM_KEY",
}
EOF
)

echo $CONFIG_JSON
printf "$CONFIG_JSON" > cfg.json

# "$@" to pass through all arguments to script
./webservice "$@"
