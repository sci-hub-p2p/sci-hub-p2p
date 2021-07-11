package main

import (
	"fmt"
	"os"

	"sci_hub_p2p/torrent"
)

func main() {
	const torrentPath = "tests/fixtures/sm_83500000-83599999.torrent"
	// content, err := os.ReadFile(torrentPath)
	// if err != nil {
	// 	return
	// }
	//
	// data, err := bencode1.Unmarshal(content)
	// if err != nil {
	// 	return
	// }

	file, err := os.Open(torrentPath)

	if err != nil {
		return
	}

	t, err := torrent.ParseReader(file)

	if err != nil {
		fmt.Println("error:", err)
		return
	}
	
	fmt.Println(t)

}
