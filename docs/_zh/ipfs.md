# IPFS

IPFS 相对 BT 有着一定的先进性，比如以文件为单位进行共享等。

但是由于 ipfs 客户端的各种问题，现有的 scimag 想要同时为 IPFS 网络提供数据是比较困难的，这需要数据保存者将 scimag 进行解压，然后将解压后的文件添加到 IPFS 网络中。

所以本项目同时提供了一个工具，如果你在使用 BT 客户端做种的话，可以同时使用这个工具来为 IPFS 网络提供数据，**而不需要解压 zip 文件。**

这样，在 IPFS 网络中，可以直接通过 CID 来获取到 PDF 文件。

使用方法也非常简单，像 ipfs 客户端一样直接添加 zip 文件即可。

```bash
./sci-hub ipfs add /path/to/zip/files/*.zip
```

<!-- prettier-ignore -->
!!! warning
    注意，这里用的不是ipfs客户端，而是本程序的 `ipfs` 子命令

如果遇到了`Too many arguments`的错误，请用`--glob`参数

```bash
./sci-hub ipfs add --glob '/path/to/zip/files/*.zip'
```

然后启动节点，程序将以 ipfs 节点的模式工作。

```bash
./sci-hub daemon start
```

本命令同时还有一个`--cache`的参数，可以指定缓存多少硬盘数据在内存中，以 MB 为单位，默认为 512。
