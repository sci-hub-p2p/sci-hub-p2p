<!-- prettier-ignore -->
!!! note
    Make sure you have downloaded all indexes files and torrents already.

All data will be stored in `~/.sci-hub-p2p/` directory, there is not way to configure it yet.

## Load torrents

Extract all torrents to `./torrents` directory.

Then let load it into database:

```bash
sci-hub torrent load ./torrents/*.torrent
```

if you met the error: `Too many arguments`, try the `--glob` flag:

```bash
sci-hub torrent load --glob './torrents/*.torrent'
```

## Load indexes

Extract all indexes to `./indexes` directory, and load them:

```bash
sci-hub indexes load ./indexes/*.jsonlines.lzma
```

if you met the error: `Too many arguments`, try the `--glob` flag:

```bash
sci-hub indexes load --glob './indexes/*.jsonlines.lzma'
```

<!-- prettier-ignore -->
!!! warning
    this will take about 30 mins and 17GB disk space.

## Fetch a paper

```bash
sci-hub paper fetch --doi '10.1145/1327452.1327492' -o map-reduce.pdf
```

You will see a new pdf file at `./map-reduce.pdf` and a CID verify

```
start downloading
expected CID: bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde
received CID: bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde
```

### About CID

you can verify it with ipfs client:

```bash
ipfs add -Q --raw-leaves \
    --hash=blake2b-256 \ # the hash we are using
    --only-hash \ # do not add, just generate the CID
    ./map-reduce.pdf
```

you should be able to see a save CID in the output:

```text
bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde
```

which is exactly the same as the CID we saw in the output.

<!-- prettier-ignore -->
!!! warning
    make sure that you are using hash `blake2b-256` and `--raw-leaves`

We have generate CID for all PDF files when indexing them,
bot not like `ipfs add` will do, we just hash them, not seeding them.
So there is a chance that you can access the PDF file through a IPFS gateway, but it's very unlikely.

This CID is generated from the raw file content, without it's filename.

So you need to download it with a filename query (again, unlikely at now)

```
https://ipfs.io/ipfs/bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde?filename=map-reduce.pdf
```

[mode details about IPFS](./ipfs.md)
