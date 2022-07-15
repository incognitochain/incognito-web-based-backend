#!/bin/sh
REDISHOST="$(getent hosts redis | awk '{ print $1 }')"

JSON='{"apiport":'"$PORT"',"chaindata":"chain","concurrentotacheck":10,"mode":"'"$MODE"'","mongo":"mongodb://root:example@'"$MONGOHOST"':27017","mongodb":"coin","chainnetwork":"'"$CHAINNETWORK"'","indexerid": '"$INDEXERID"',"masterindexer":"'"$INDEXERADDR"':9009","analyticsAPIEndpoint": "'"$ANALYTICS"'","externaldecimals":"'"$EXDECIMALS"'","fullnodedata":'"$FULLNODEDATA"',"coordinator":"'"$COORIDANTORADDR"':9009","logrecorder":"'"$LOGRECORDER"':9008"}'
echo $JSON
echo $JSON > cfg.json
./coinservice