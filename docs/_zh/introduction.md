---
permalink: 简介
---

# 简介

[创世纪图书馆](https://zh.wikipedia.org/zh-cn/%E5%88%9B%E4%B8%96%E7%BA%AA%E5%9B%BE%E4%B9%A6%E9%A6%86)（Library Genesis，或缩写为 LibGen) 是真正储存 Sci-Hub 论文的地方。LibGen 打包了所有其储存的学术论文（称为 scimag ），并且提供了[全部的种子文件](https://libgen.rs/scimag/repository_torrent/)。

Reddit 上发起了许多[数据保存项目](https://www.reddit.com/r/DataHoarder/comments/nc27fv/rescue_mission_for_scihub_and_open_science_we_are/)，号召大家使用 BT 保存所有 LibGen 储存的论文。

本项目包含两个部分:

## 一个 P2P 客户端，从 P2P 网络中获取论文

提供 SciHub 网站相同的功能，通过 DOI 从 P2P 网络中获取论文，而不再需要一个额外的 DNS 或者 HTTP 服务器。

## 一个 IPFS 节点客户端，在 IPFS 网络中提供 scimag 备份的 PDF 文件

IPFS 网络跟 BT 网络一样，数据储存在所有的 P2P 节点。

本项目为 scimag 的数据保存者提供了一个工具，可以帮助所有的数据提供者同时在 IPFS 项目中提供 PDF 文件。

也就是说，如果您在为 scimag 做种，使用本工具，在几乎不需要额外的硬盘空间的情况下，可以在 IPFS 网络中同时提供您保存的论文文件。

<!-- prettier-ignore -->
!!! note
    "几乎不需要额外的硬盘空间"意味着，每一个种子的数据（约100GB）需要额外消耗大约200MB的空间。

更多信息，请参考 [IPFS 的相关 文档](./ipfs.md)。

## 项目状态

如果你对开发进展感兴趣，可以察看[此 issue](https://github.com/sci-hub-p2p/sci-hub-p2p/issues/2)
