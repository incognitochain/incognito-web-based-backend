#!/bin/sh
MONGOHOST="$(getent hosts mongo | awk '{ print $1 }')"

JSON='{"Port":'"$Port"',"NetworkID": "'"$NetworkID"'","Mode":"'"$MODE"'","Mongo":"mongodb://'"${MONGO_USERNAME}"':'"${MONGO_PASSWORD}"'@'"$MONGOHOST"':27017","Mongodb":"data","CoinserviceURL":"'"$CoinserviceURL"'","FullnodeURL":"'"$FullnodeURL"'","ShieldService":"'"$ShieldService"'","CaptchaSecret":"'"$CaptchaSecret"'"}'
echo $JSON
echo $JSON > cfg.json
./webservice