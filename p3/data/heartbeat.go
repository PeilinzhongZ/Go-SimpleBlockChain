package data

import (
// "../../p1"
// "math/rand"
// "time"
)

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Addr        string `json:"addr"`
	Hops        int32  `json:"hops"`
}

func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	return HeartBeatData{IfNewBlock: ifNewBlock, Id: id, BlockJson: blockJson, PeerMapJson: peerMapJson, Addr: addr}
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJson string, addr string) HeartBeatData {
	// rand.Seed(time.Now().UnixNano())
	// if rand.Intn(2) == 1 {
	// 	var mpt p1.MerklePatriciaTrie
	// 	mpt.Initial()
	// 	mpt.Insert("aaa", "apple")
	// 	// mpt.Insert("aap", "banana")
	// 	// mpt.Insert("bb", "right leaf")
	// 	block := sbc.GenBlock(mpt)
	// 	heartBeatData := NewHeartBeatData(true, selfId, block.Encode(), peerMapJson, addr)
	// 	heartBeatData.Hops = 2
	// 	return heartBeatData
	// }
	heartBeatData := NewHeartBeatData(false, selfId, "", peerMapJson, addr)
	heartBeatData.Hops = 2
	return heartBeatData
}
