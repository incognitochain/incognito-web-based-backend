#!/bin/sh
REDISHOST_FROM_HOSTNAME="$(getent hosts redisstack | awk '{ print $1 }')"
REDISHOST="${$REDISHOST:-REDISHOST_FROM_HOSTNAME}"

CONFIG_JSON=$(cat <<EOF
{
  "Port": $Port,
  "NetworkID": "$NetworkID",
  "Mode": "$MODE",
  "Database":["$REDISHOST:6379"],
  "DBUSER":"$DBUser",
  "DBPASS":"$DBPass",
  "CoinserviceURL": "$CoinserviceURL",
  "FullnodeURL": "$FullnodeURL",
  "ShieldService": "$ShieldService",
  "FaucetService": "$FaucetService",
  "CaptchaSecret":"$CaptchaSecret"
}
EOF
)

echo $CONFIG_JSON
printf "$CONFIG_JSON" > cfg.json

# "$@" to pass through all arguments to script
./webservice "$@"
