#!/usr/bin/env bash

set -ex

#go build -o ./dist/sci-hub.exe

#./dist/sci-hub.exe indexes gen -t /d/data/sm_11200000-11299999.torrent -d /d/data/ --parallel=4 --disable-progress
#
#./dist/sci-hub.exe indexes load ./out/*.lzma
#
#./dist/sci-hub.exe torrent load /d/data/*.torrent

./dist/sci-hub.exe paper fetch --doi '10.1145/1327452.1327492' -o map-reduce.pdf
