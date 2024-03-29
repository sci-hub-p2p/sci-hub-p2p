# 索引

索引是从每个种子下载后的完整数据包中生成的，如果将来 LibGen 打包了将新的论文打包，发布了新的种子，那么我们需要针对对应的文件生成新的索引。

<!-- prettier-ignore-->
!!! info "提示"
    索引已经生成完成，使用说明请参考 [使用说明](../usage.md)

## 生成

```bash
sci-hub indexes gen -t /path/to/data.torrent -d /path/to/download/dir/
```

```text
/path/to/download/dir/
|-- 11200000
|-- ...
`-- 55900000
```

`/path/to/download/dir/` 应该是在下载工具中设置的下载链接，而不是下载完成的种子文件夹。

你能看到一个进度条提示

```bash
./sci-hub indexes gen -t ~/torrents/sm_55900000-55999999.torrent -d ~/data/
# start generate indexes
# 22879 / 100000 [--------->_________________________________] 22.88% 1607 p/s
```

这个命令会根据种子和 zip 文件的内容生成一个索引，储存在`./out/`文件夹里。原始的索引文件 `{info hash}.indexes`，以及用于传输和导入的 `{info hash}.jsonlines.lzma`。

```console
$ ls out/
2afe5336ccf75d633fc7aac7c95342556745ad39.indexes
2afe5336ccf75d633fc7aac7c95342556745ad39.jsonlines.lzma
```

一个批量生成的 bash 脚本:

```bash
#!/usr/bin/env bash

FILES="$HOME/repository_torrent/sm_*.torrent"
for f in $FILES; do
  echo "Processing $f file..."
  ./sci-hub indexes gen -t "$f" -d "/path/to/download/dir"
done
```
