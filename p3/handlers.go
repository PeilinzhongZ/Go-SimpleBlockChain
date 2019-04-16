package p3

import (
	"../p1"
	"../p2"
	"./data"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	// "github.com/gorilla/mux"
	// "io"
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "http://localhost:6687"
var FIRST_ADDR = "http://localhost:6686"

var PEERS_SIZE = int32(32)

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool

func initial() {
	// This function will be executed before everything else.
	// Do some initialization here.
	SBC = data.NewBlockChain()
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	if !ifStarted {
		if len(os.Args) > 1 {
			SELF_ADDR = "http://localhost:" + os.Args[1]
		}
		initial()
		id64, err := strconv.ParseInt(string(os.Args[1]), 10, 32)
		id := int32(id64)
		// id, err := Register()
		if err != nil {
			fmt.Fprintf(w, "Register error")
			return
		}
		Peers = data.NewPeerList(id, PEERS_SIZE)
		first, ok := r.URL.Query()["first"]
		if ok && first[0] == "true" {
			var gBlock p2.Block
			gBlock.FirstBlock()
			SBC.Insert(gBlock)
		} else {
			hearbeatData, err := GetPeerMap()
			if err != nil {
				fmt.Fprintf(w, "GetPeerMap error")
				return
			}
			Peers.Add(hearbeatData.Addr, hearbeatData.Id)
			Peers.InjectPeerMapJson(hearbeatData.PeerMapJson, SELF_ADDR)
			blockChainJSON, err := Download()
			if err != nil {
				fmt.Fprintf(w, "Donwload error")
				return
			}
			SBC.UpdateEntireBlockChain(blockChainJSON)
		}
		rand.Seed(time.Now().UnixNano())
		StartHeartBeat()
		StartTryingNonces()
		ifStarted = true
	}
}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to TA's server, get an ID
func Register() (int32, error) {
	resp, err := http.Get(REGISTER_SERVER)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(string(body), 10, 32)
	return int32(i), err
}

func GetPeerMap() (data.HeartBeatData, error) {
	resp, err := http.Get(FIRST_ADDR + "/peerMap?addr=" + SELF_ADDR + "&id=" + strconv.Itoa(int(Peers.GetSelfId())))
	if err != nil {
		return data.HeartBeatData{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data.HeartBeatData{}, err
	}
	var heartBeat data.HeartBeatData
	if err := json.Unmarshal(body, &heartBeat); err != nil {
		return data.HeartBeatData{}, err
	}
	return heartBeat, err
}

// Download blockchain from TA server
func Download() (string, error) {
	peers := Peers.Copy()
	for addr := range peers {
		resp, err := http.Get(addr + "/upload")
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		return string(body), err
	}
	return "", errors.New("error")
}

func UploadPeerMap(w http.ResponseWriter, r *http.Request) {
	addr, ok1 := r.URL.Query()["addr"]
	idString, ok2 := r.URL.Query()["id"]
	id, err := strconv.Atoi(idString[0])
	if ok1 && ok2 && err == nil {
		Peers.Add(addr[0], int32(id))
		peerMapJSON, err := Peers.PeerMapToJson()
		if err != nil {
			// handle error
		}
		heartbeat := data.HeartBeatData{Id: Peers.GetSelfId(), PeerMapJson: peerMapJSON, Addr: SELF_ADDR}
		heartbeatJSON, err := json.Marshal(heartbeat)
		if err != nil {
			// handle error
		}
		fmt.Fprint(w, string(heartbeatJSON))
	}
}

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		return
	}
	fmt.Fprint(w, blockChainJson)
}

// Upload a block to whoever called this method, return jsonStr
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	i, err := strconv.ParseInt(path[2], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		block, ok := SBC.GetBlock(int32(i), path[3])
		if ok == false {
			w.WriteHeader(http.StatusNoContent)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, block.Encode())
	}
}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		// handle error
		return
	}
	var heartBeat data.HeartBeatData
	if err := json.Unmarshal(body, &heartBeat); err != nil {
		// handle error
		return
	}
	Peers.Add(heartBeat.Addr, heartBeat.Id)
	Peers.InjectPeerMapJson(heartBeat.PeerMapJson, SELF_ADDR)
	if heartBeat.IfNewBlock {
		var block p2.Block
		block.Decode(heartBeat.BlockJson)
		if exist := SBC.CheckParentHash(block); exist {
			ok := CheckNonce(block)
			if ok {
				SBC.Insert(block)
			}
		} else {
			AskForBlock(block.Header.Height-1, block.Header.ParentHash)
			if exist = SBC.CheckParentHash(block); exist {
				ok := CheckNonce(block)
				if ok {
					SBC.Insert(block)
				}
			}
		}
		if heartBeat.Hops = heartBeat.Hops - 1; heartBeat.Hops != 0 {
			heartBeat.Addr = SELF_ADDR
			heartBeat.Id = Peers.GetSelfId()
			ForwardHeartBeat(heartBeat)
		}
	}
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	peers := Peers.Copy()
	for addr := range peers {
		resp, err := http.Get(addr + "/block/" + strconv.Itoa(int(height)) + "/" + hash)
		if err != nil {
			// handle error
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				// handle error
				continue
			}
			var block p2.Block
			block.Decode(string(body))
			if exist := SBC.CheckParentHash(block); !exist {
				AskForBlock(block.Header.Height-1, block.Header.ParentHash)
				if exist = SBC.CheckParentHash(block); exist {
					// ok := CheckNonce(block)
					// if ok {
					SBC.Insert(block)
					// }
				}
			} else {
				// ok := CheckNonce(block)
				// if ok {
				SBC.Insert(block)
				// }
			}
			break
		}
	}
}

func CheckNonce(block p2.Block) bool {
	str := sha3.Sum256([]byte(block.Header.ParentHash + block.Header.Nonce + block.Value.Root))
	result := hex.EncodeToString(str[:])
	return strings.HasPrefix(result, "000000")
}

func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	list := Peers.Copy()
	for addr := range list {
		go func(addr string) {
			heartBeatJSON, err := json.Marshal(heartBeatData)
			if err != nil {
				// handle error
			}
			body := bytes.NewBuffer(heartBeatJSON)
			_, err = http.Post(addr+"/heartbeat/receive", "application/json", body)
			if err != nil {
				// handle error
			}
		}(addr)
	}
}

func StartHeartBeat() {
	go func() {
		for range time.Tick(time.Second * 10) {
			list := Peers.Copy()
			str, err := Peers.PeerMapToJson()
			if err != nil {
				str = ""
			}
			heartBeatData := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), str, SELF_ADDR)
			heartBeatJSON, err := json.Marshal(heartBeatData)
			if err != nil {
				// handle error
			}
			body := bytes.NewBuffer(heartBeatJSON)
			for addr := range list {
				go func(addr string) {
					_, err = http.Post(addr+"/heartbeat/receive", "application/json", body)
					if err != nil {
						// handle error
						Peers.Delete(addr)
					}
				}(addr)
			}
		}
	}()
}

func StartTryingNonces() {
	go func() {
		for true {
			latestBlocks, ok := SBC.GetLatestBlocks()
			if ok {
				var mpt p1.MerklePatriciaTrie
				mpt.Initial()
				mpt.Insert("aaa", "apple")
				x, success := TryNonces(latestBlocks, mpt.Root)
				if success {
					str, err := Peers.PeerMapToJson()
					if err != nil {
						str = ""
					}
					var block p2.Block
					block.Initial(latestBlocks[0].Header.Height+1, latestBlocks[0].Header.Hash, mpt)
					block.Header.Nonce = x
					SBC.Insert(block)
					heartBeadData := data.NewHeartBeatData(true, Peers.GetSelfId(), block.Encode(), str, SELF_ADDR)
					heartBeadData.Hops = 2
					fmt.Println("------", block.Header.Height, "------")
					ForwardHeartBeat(heartBeadData)
				}
			}
		}
	}()
}

func TryNonces(latestBlocks []p2.Block, Root string) (string, bool) {
	var result string
	var x string
	success := true
	for !strings.HasPrefix(result, "000000") {
		if SBC.GetLength() > latestBlocks[0].Header.Height {
			success = false
			break
		}
		bytes := make([]byte, 8)
		rand.Read(bytes)
		x = hex.EncodeToString(bytes)
		resultSum := sha3.Sum256([]byte(latestBlocks[0].Header.Hash + x + Root))
		result = hex.EncodeToString(resultSum[:])
	}
	return x, success
}

func Canonical(w http.ResponseWriter, r *http.Request) {
	lastBlocks, ok := SBC.GetLatestBlocks()
	if !ok {
		//handle error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var chains string
	for i, block := range lastBlocks {
		chain := "Chain" + strconv.Itoa(i) + "\n"
		chain += block.Info()
		for block.Header.Height != 0 {
			block, _ = SBC.GetParentBlock(block)
			chain += block.Info()
		}
		chains += chain
	}
	fmt.Fprint(w, chains)
}
