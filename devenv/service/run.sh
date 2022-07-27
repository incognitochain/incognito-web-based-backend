#!/bin/sh
REDISHOST="$(getent hosts redisstack | awk '{ print $1 }')"

JSON='{"Port":'"$Port"',"NetworkID": "'"$NetworkID"'","Mode":"'"$MODE"'","DatabaseURLs":["'"$REDISHOST"':6379"],"DBUSER":"'"$DBUser"'","DBPASS":"'"$DBPass"'","CoinserviceURL":"'"$CoinserviceURL"'","FullnodeURL":"'"$FullnodeURL"'","ShieldService":"'"$ShieldService"'","CaptchaSecret":"'"$CaptchaSecret"'"}'
echo $JSON
echo $JSON > cfg.json
./webservice