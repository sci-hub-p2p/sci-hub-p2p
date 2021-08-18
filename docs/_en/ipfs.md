# IPFS

I have create a tool to serve PDF files to IPFS. If you are seeding the torrents,
please DO USE IT.

This tools could serve PDF files directly from the sci-mag archive, without need to decompress them.

and it's easy to use:

```bash
./sci-hub ipfs add /path/to/zip/files/*.zip
```

If you meed `Arguments Too Long`, use `--glob` flag:

```bash
./sci-hub ipfs add --glob '/path/to/zip/files/*.zip'
```

then start your node:

```bash
./sci-hub daemon start # --cache 512
```

there is a cache flag that will will sci-hub-p2p how memory it will use to cache the data, avoiding read from dist too much.
