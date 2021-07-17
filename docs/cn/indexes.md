# 索引

## 生成

```console
$ ./sci-hub indexes gen -t /path/to/data.torrent -d /path/to/data/dir/ --parallel 8
```

```text
/path/to/data/
`-- 55900000
    |-- libgen.scimag55900000-55900999.zip
    |-- libgen.scimag55901000-55901999.zip
    ...
    |-- libgen.scimag55997000-55997999.zip
    |-- libgen.scimag55998000-55998999.zip
    `-- libgen.scimag55999000-55999999.zip

1 directory, 100 files
```

`/path/to/data/` 应该是在下载工具中设置的下载链接，而不是下载完成的种子文件夹。

你能看到一个进度条提示

```console
$ ./sci-hub indexes gen -t ~/data/sm_55900000-55999999.torrent -d ~/data/
start generate indexes
22879 / 100000 [-------------->________________________________________________] 22.88% 1607 p/s
```

这个命令会根据种子和 zip 文件的内容生成一个索引，储存在`./out/`文件夹里。原始的索引文件 `{info hash}.indexes`，以及 gzip 压缩过的 `{info hash}.indexes.lzma`。

```console
$ ls out/
2afe5336ccf75d633fc7aac7c95342556745ad39.indexes
2afe5336ccf75d633fc7aac7c95342556745ad39.indexes.lzma
```
