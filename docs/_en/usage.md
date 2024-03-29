## Prerequisites

In order to complete this guide, you will first need to perform the following tasks on your computer.

### Create a folder

Create a working folder for this project, we would use `~/sci-hub/` in this guide.

### Download Sci-hub-p2p

Download the latest release from [GitHub Release](https://github.com/sci-hub-p2p/sci-hub-p2p/releases), and put it under `~/sci-hub/`.

### Download SciMag torrent files

To set this up, download all torrent files [here](https://libgen.rs/scimag/repository_torrent/) and put them in `~/sci-hub/torrents/`.

### Download index files

Download index files [here](https://github.com/sci-hub-p2p/artifacts/releases/tag/0). Make sure index files are correctly located under `~/sci-hub/index`.

## Data storage location

You can set environment variable `APP_HOME` to configure the location where to store all data.

Is this environment variable is not set, the tool will use `~/.sci-hub-p2p/` as default location.

## Load torrents

To import all torrent seeds under `~/.sci-hub/torrents/`, run:

```bash
cd ~/sci-hub/
./sci-hub torrent load --glob '~/sci-hub/torrents/*.torrent'
```

This process would only take a few seconds.

## Load indices

To load all indices to database, run:

```bash
./sci-hub indexes load --glob '~/sci-hub/index/*.lzma'
```

<!-- prettier-ignore -->
!!! warning
    The whole process could take about 30 minutes or longer, make sure you have ~20G of hard disk space under your home folder (`~/.sci-hub-p2p/`).

## Fetch a paper

Now, you would be able to get any papers exist in SciMag Collection.

Let's take Google's MapReduce paper as an example, run:

```bash
./sci-hub paper fetch --doi '10.1145/1327452.1327492' -o ./map-reduce.pdf
```

Use -o to specify the output path.

```text
#Output

start downloading
expected CID: bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde
received CID: bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde
```

You could find the CID of this paper, which is used to verify the integrity of papers.

If you would like to use IPFS, [see here](./ipfs.md).
