package p2

import (
	"../p1"
	"./format"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"time"
)

type Block struct {
	Header header
	Value  p1.MerklePatriciaTrie
}

type header struct {
	Height     int32
	Timestamp  int64
	Hash       string
	ParentHash string
	Size       int32
	Nonce      string
}

func (block *Block) Initial(height int32, parentHash string, mpt p1.MerklePatriciaTrie) {
	block.Header.Height = height
	block.Header.ParentHash = parentHash
	block.Header.Timestamp = time.Now().Unix()
	block.Value = mpt
	block.Header.Size = int32(len([]byte(mpt.String())))
	block.Header.Hash = block.hashMpt(height, block.Header.Size, block.Header.Timestamp, parentHash, mpt.Root)
}

func (block *Block) FirstBlock() {
	block.Header.Height = 0
	block.Header.ParentHash = ""
	block.Header.Timestamp = time.Now().Unix()
	block.Header.Size = 0
	block.Header.Hash = "GENESIS"
}

func (block *Block) hashMpt(height, size int32, timestamp int64, parentHash, root string) string {
	var str string
	str = string(height) + string(timestamp) + parentHash + root + string(size)
	sum := sha3.Sum256([]byte(str))
	return hex.EncodeToString(sum[:])
}

func (block *Block) Encode() string {
	blockData := format.BlockData{}
	blockData.Inital(block.Header.Height, block.Header.Timestamp,
		block.Header.Hash, block.Header.ParentHash, block.Header.Size, block.Value.Values, block.Header.Nonce)
	blockString, _ := json.Marshal(blockData)
	return string(blockString)
}

func (block *Block) Decode(jsonData string) {
	var blockData format.BlockData
	if err := json.Unmarshal([]byte(jsonData), &blockData); err != nil {
		panic(err)
	}
	mpt := p1.MerklePatriciaTrie{}
	for k, v := range blockData.Mpt {
		mpt.Insert(k, v)
	}
	block.Header.Height = blockData.Height
	block.Header.Timestamp = blockData.Timestamp
	block.Header.Hash = blockData.Hash
	block.Header.ParentHash = blockData.ParentHash
	block.Header.Size = blockData.Size
	block.Value = mpt
	block.Header.Nonce = blockData.Nonce
}

func (block *Block) Info() string {
	return fmt.Sprintf("height=%v, timestamp=%v, hash=%s, parentHash=%s\n",
		block.Header.Height, block.Header.Timestamp, block.Header.Hash, block.Header.ParentHash)
}
