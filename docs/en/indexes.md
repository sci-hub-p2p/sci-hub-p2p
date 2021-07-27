# Index

## Generate

```console
$ . /sci-hub indexes gen -t /path/to/data.torrent -d /path/to/download/dir/ --parallel 4
```

```text
/path/to/download/dir/
|-- 11200000
|-- ...
`-- 55900000
```

`/path/to/download/dir/` should be the download path you set in your BitTorrent client, not the dir with torrent name.

You can see a progress prompt:

```console
$ . /sci-hub indexes gen -t ~/torrents/sm_55900000-55999999.torrent -d ~/data/
start generate indexes
22879 / 100000 [-------------->________________________________________________] 22.88% 1607 p/s
```

This command will generate an index based on the contents of the torrent and zip files,
which will be stored in the `. /out/` folder. The original index file `{info hash}.indexes`, and `{info hash}.jsonlines.lzma` for transmission and importing.

```console
$ ls out/
2afe5336ccf75d633fc7aac7c95342556745ad39.indexes
2afe5336ccf75d633fc7aac7c95342556745ad39.jsonlines.lzma
```

Usually, `{info hash}.jsonlines.lzma` should be about 4~5MB

A batch-generating bash script:

```bash
#! /usr/bin/env bash

FILES="$HOME/repository_torrent/sm_*.torrent"
for f in $FILES; do
  echo "Processing $f file..."
  . /sci-hub indexes gen -t "$f" -d "/path/to/download/dir"
done
```
