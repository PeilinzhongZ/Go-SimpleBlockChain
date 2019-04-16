package data

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	// "strings"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

type PeerListJson struct {
	SelfId    int32
	PeerMap   map[string]int32
	MaxLength int32
}

type peer struct {
	Addr string
	ID   int32
}

func NewPeerList(id int32, maxLength int32) PeerList {
	return PeerList{selfId: id, peerMap: make(map[string]int32), maxLength: maxLength}
}

func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	peers.peerMap[addr] = id
	peers.mux.Unlock()
}

func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

func (peers *PeerList) Rebalance() {
	var list []peer
	if len(peers.peerMap) > int(peers.maxLength) {
		for k, v := range peers.peerMap {
			list = append(list, peer{k, v})
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].ID < list[j].ID
		})
		var index int
		for i, v := range list {
			if v.ID > peers.selfId {
				index = i
				break
			}
		}
		prefix := list[:index]
		sufix := list[index:]
		if len(prefix) < int(peers.maxLength)/2 {
			list = append(list[:index+int(peers.maxLength)/2], list[len(list)-int(peers.maxLength)/2+len(prefix):]...)
		} else if len(sufix) < int(peers.maxLength)/2 {
			list = append(list[index-int(peers.maxLength)/2:], list[:int(peers.maxLength)/2-len(sufix)]...)
		} else {
			list = list[index-int(peers.maxLength)/2 : index+int(peers.maxLength)/2]
		}
		peerMap := make(map[string]int32)
		for _, v := range list {
			peerMap[v.Addr] = v.ID
		}
		peers.peerMap = peerMap
	}
}

func (peers *PeerList) Show() string {
	peers.mux.Lock()
	peerListJSON := &PeerListJson{peers.selfId, peers.peerMap, peers.maxLength}
	str, _ := json.Marshal(peerListJSON)
	defer peers.mux.Unlock()
	return string(str)
}

func (peers *PeerList) Register(id int32) {
	peers.selfId = id
	fmt.Printf("SelfId=%v\n", id)
}

func (peers *PeerList) Copy() map[string]int32 {
	peers.mux.Lock()
	peers.Rebalance()
	m := make(map[string]int32)
	for k, v := range peers.peerMap {
		m[k] = v
	}
	defer peers.mux.Unlock()
	return m
}

func (peers *PeerList) GetSelfId() int32 {
	return peers.selfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {
	peers.mux.Lock()
	str, err := json.Marshal(peers.peerMap)
	defer peers.mux.Unlock()
	return string(str), err
}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	var peerMap map[string]int32
	if err := json.Unmarshal([]byte(peerMapJsonStr), &peerMap); err != nil {
		panic(err)
	}
	peers.mux.Lock()
	for k, v := range peerMap {
		if k != selfAddr {
			peers.peerMap[k] = v
		}
	}
	peers.mux.Unlock()
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
}
