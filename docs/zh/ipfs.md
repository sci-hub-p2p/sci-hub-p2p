索引中包括了所有的 论文在 IPFS 网络中的 CID，理论上说是可以通过 IPFS 网络找到对应的论文的。

但是由于 ipfs 节点数的问题，实际上想要通过 IPFS 获取到论文是非常困难的。

但是本程序同时提供了一个工具，如果你在使用 BT 客户端做种的话，请同时使用这个工具来为 IPFS 网络提供数据。

如果你对 IPFS 有一定了解的话，可能知道 IPFS 的客户端需要把 ZIP 文件解压，才能添加其中的 PDF 文件。
但是使用本工具，可以直接添加 zip 文件，在 IPFS 网络中提供的是原始的 PDF 文件，而非整个 ZIP 文件。

这样，在 IPFS 网络中，可以直接通过 CID 来获取到 PDF 文件。

使用方法也非常简单，像 ipfs 一样直接添加 zip 文件即可。

```bash
./sci-hub ipfs add /path/to/zip/files/*.zip
```

如果遇到了`Arguments Too Long`的错误，请用`--glob`参数

```bash
./sci-hub ipfs add --glob '/path/to/zip/files/*.zip'
```

然后启动节点

```bash
./sci-hub daemon start
```
