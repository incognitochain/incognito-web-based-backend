#!/bin/sh
REDISHOST_FROM_HOSTNAME="$(getent hosts redisstack | awk '{ print $1 }')"
REDIS_HOST="${REDIS_HOST:-$REDISHOST_FROM_HOSTNAME}"
REDIS_PORT="${REDIS_PORT:-6379}"

CONFIG_JSON=$(cat <<EOF
{
  "Port": $PORT,
  "NetworkID": "$NETWORK_ID",
  "Mode": "$MODE",
  "DatabaseURLs":["$REDIS_HOST:$REDIS_PORT"],
  "DBUSER":"$DB_USER",
  "DBPASS":"$DB_PASSWORD",
  "CoinserviceURL": "$COIN_SERVICE_URL",
  "FullnodeURL": "$FULLNODE_URL",
  "ShieldService": "$SHIELD_SERVICE_URL",
  "FaucetService": "$FAUCET_SERVICE_URL",
  "CaptchaSecret":"$CAPTCHA_SECRET"
}
EOF
)

echo $CONFIG_JSON
printf "$CONFIG_JSON" > cfg.json

# "$@" to pass through all arguments to script
./webservice "$@"
