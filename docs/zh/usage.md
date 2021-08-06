在开始使用之前，请先:

1. 从 [GitHub Release](https://github.com/sci-hub-p2p/sci-hub-p2p/releases)下载最新的源程序

2. 使用 BT 下载最新的索引文件，种子可以在这里找到：[sci-hub-p2p/artifacts](https://github.com/sci-hub-p2p/artifacts/releases/tag/0)

3. 下载所有的种子 <https://libgen.rs/scimag/repository_torrent/>

## 数据的保存目录

可以通过环境变量 `APP_HOME` 来设置程序所有持久化数据的地址。

如果没有此环境变量，所有的数据会保存在 `~/.sci-hub-p2p/` 文件夹中

## 导入索引

首先解压索引文件到任意文件夹，这里以 `/path/to/indexes/` 为例。

```bash
./sci-hub indexes load --glob '/path/to/indexes/*.lzma'
```

整个过程大概会需要 30 分钟，占用约 17G 的硬盘空间。

## 导入种子

然后导入全部的种子文件，以`/path/to/torrents/*.torrent`为例:

```bash
./sci-hub torrent load --glob '/path/to/torrents/*.torrent'
```

这个过程大概只要几秒就可以完成。

## 获取论文

现在，就可以获取任意数据库中存在的论文了。

以 google 的 map reduce 论文为例，以`-o`指定输出路径

```bash
./sci-hub paper fetch --doi '10.1145/1327452.1327492' -o ./map-reduce.pdf
```

应该会看到这样的输出

```text
start downloading
expected CID: bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde
received CID: bafk2bzaceav734ba4n55d24e4ihka74oeuo42uwmh5a2dryiivcprt2ga3zde
```

这是这篇论文的 CID，用来验证数据正确性。

关于更多 IPFS 的内容，见 [这里](./ipfs.md)。
