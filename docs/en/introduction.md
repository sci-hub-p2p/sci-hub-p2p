Library Genesis (LibGen) is the largest free library in history. Volunteers and contributors of [this data holding project](https://www.reddit.com/r/libgen/comments/eo0y2c/library_genesis_project_update_25_million_books/) collect, pack and distribute books and scientific articles with Peer to Peer (P2P) network. SciMag Collection is one of those data holding projects, it contains more than 85M papers collected by Sci-Hub.

Our project is based on SciMag Collection, BitTorrent and the IPFS network, aiming to provide the same experience as the original [Sci-Hub website](https://sci-hub.st/), retrieval papers from the P2P network by DOI while not relying on any DNS or HTTP server.

<!-- prettier-ignore -->
!!! warning
    Before you begin, make sure you understand [the legal implications of hosting and sharing copyrighted material](https://www.nolo.com/legal-encyclopedia/what-to-do-if-your-named-bit-torrent-lawsuit.html).

    If your ISP does not allow BitTorrent traffic or you're not sure, **DO NOT USE IT!**

## Project Composition

This project consists of two parts:

### A P2P client

Get papers from P2P network through DOI, similar to SciHub website. No additional DNS or HTTP server is needed.

### An IPFS node client

This project provides a tool that allows SciMag seeders on BT networks to seed directly on IPFS networks without unzipping archived packages.

In other words, if your are seeding SciMag Collection and using this tool at the same time, you can seed on the IPFS network with almost no additional hard disk space.

<!-- prettier-ignore -->
!!!note
    "Almost no additional hard disk space" means that each seed data (about 100GB) needs to consume about 200MB of additional space.

For more information, please refer to the relevant documentation of [IPFS](./ipfs.md).

## Project Status

If you are interested in development progress, you can check [this issue](https://github.com/sci-hub-p2p/sci-hub-p2p/issues/2)
