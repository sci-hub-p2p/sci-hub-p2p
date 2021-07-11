package torrent

import "fmt"

type Node struct {
	host string
	port int
}

type Torrent struct {
	info
	Announce     string
	AnnounceList [][]string
	CreationDate int
	Nodes        []Node `json:"nodes"`
	InfoHash     string
}

func (t Torrent) String() string {
	return fmt.Sprintf("Torrent{Name=%s, info_hash=%s}", t.Name, t.InfoHash)
}
