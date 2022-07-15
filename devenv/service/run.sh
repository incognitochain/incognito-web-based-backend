#!/bin/sh
REDISHOST="$(getent hosts redis | awk '{ print $1 }')"

JSON='{"Port":'"$PORT"',"DatabaseURLs":["'"$REDISHOST"'"],"CoinserviceURL":"'"$CoinserviceURL"'","FullnodeURL":"'"$FullnodeURL"'","ShieldService":"'"$ShieldService"'"}'
echo $JSON
echo $JSON > cfg.json
./webservice