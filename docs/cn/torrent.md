# 种子

## 把种子导入数据库

```console
$ ./sci-hub torrent load ./repository_torrent/sm_*.torrent
```

如果你遇到了"Argument list too long"的错误，可以使用 glob 来导入种子:

```console
$ ./sci-hub torrent load --glob "~/repository_torrent/sm_*.torrent"
```

也可以一起用:

```console
$ ./sci-hub torrent load ./a.torrent --glob "$HOME/repository_torrent/sm_*.torrent"
```
