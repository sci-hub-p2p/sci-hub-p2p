Library Genesis (LibGen) is the largest free library in history. Volunteers and contributors of [this data holding project](https://www.reddit.com/r/libgen/comments/eo0y2c/library_genesis_project_update_25_million_books/) collect, pack and distribute books and scientific articles with Peer to Peer (P2P) network. SciMag Collection is one of those data preservation projects, it stores all 85M papers on Sci-Hub. 

Our project is based on SciMag Collection, P2P and the IPFS networks, aiming to provide same experience as the original [Sci-Hub website](https://sci-hub.st/),retrieval papers from the P2P network by DOI while not relying on any web or DNS server.

<!-- prettier-ignore -->
!!! warning
    Before you begin, make sure you understand [the legal implications of hosting and sharing copyrighted material](https://www.nolo.com/legal-encyclopedia/what-to-do-if-your-named-bit-torrent-lawsuit.html).

    If your ISP does not allow BitTorrent traffic or you're not sure, **DO NOT USE IT!**


This project consists of two parts:

### A P2P client 

Get papers from P2P network, similar to SciHub website, get papers from P2P network through DOI, but no additional DNS or HTTP server is needed.

### An IPFS node client

The IPFS network is the same as the BT network, where data is stored in P2P nodes. This project provides a tool for SciMag seeders to seed PDF files in the IPFS network at the same time.

In other words, if your are torrenting SciMag Collection and using this tool at the same time, you can seed on the IPFS network at the same time with almost no additional hard disk space.

Note

"Almost no additional hard disk space" means that each seed data (about 100GB) needs to consume about 200MB of additional space.

For more information, please refer to the relevant documentation of IPFS.

project status
If you are interested in development progress, you can check this issue
