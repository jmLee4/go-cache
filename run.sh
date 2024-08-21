#!/bin/bash
trap "rm server; kill 0" EXIT

go build -o server
./server --cacheServerPort=8001 &
./server --cacheServerPort=8002 &
./server --cacheServerPort=8003 --startAPIServerFlag=1 &

sleep 2
echo ">>> start test"
curl --noproxy "localhost" -w "\n" "http://localhost:9999/api?key=Tom" &
curl --noproxy "localhost" -w "\n" "http://localhost:9999/api?key=Tom" &
curl --noproxy "localhost" -w "\n" "http://localhost:9999/api?key=Tom" &

wait