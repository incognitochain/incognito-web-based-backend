#!/bin/sh
REDISHOST="$(getent hosts redis | awk '{ print $1 }')"

JSON='{"Port":'"$Port"',"DatabaseURLs":["'"$REDISHOST"':6379"],"CoinserviceURL":"'"$CoinserviceURL"'","FullnodeURL":"'"$FullnodeURL"'","ShieldService":"'"$ShieldService"'"}'
echo $JSON
echo $JSON > cfg.json
./webservice