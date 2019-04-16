package p2

import (
	"../p1"
	"./format"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"sort"
	"time"
)

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func NewBlockChain() BlockChain {
	return BlockChain{Chain: make(map[int32][]Block)}
}

func (chain *BlockChain) Get(height int32) ([]Block, bool) {
	if height > chain.Length {
		return nil, false
	}
	value, ok := chain.Chain[height]
	if ok {
		return value, ok
	}
	return nil, ok
}

func (chain *BlockChain) Insert(block Block) {
	// if chain.Chain == nil {
	// 	chain.Chain = make(map[int32][]Block)
	// }
	blocks := chain.Chain[block.Header.Height]
	if len(blocks) == 0 {
		blocks = []Block{block}
		chain.Chain[block.Header.Height] = blocks
		chain.Length = block.Header.Height
	} else {
		var t bool
		for _, v := range blocks {
			if v.Header.Hash == block.Header.Hash {
				t = true
			}
		}
		if t == false {
			blocks = append(blocks, block)
			chain.Chain[block.Header.Height] = blocks
		}
	}
}

func (chain *BlockChain) GenBlock(mpt p1.MerklePatriciaTrie) Block {
	var block Block
	parentHashs, bool := chain.Get(chain.Length)
	if bool {
		rand.Seed(time.Now().UnixNano())
		i := rand.Intn(len(parentHashs))
		block.Initial(chain.Length+1, parentHashs[i].Header.Hash, mpt)
		chain.Insert(block)
	}
	return block
}

func (chain *BlockChain) Encode() (string, error) {
	var blockList []format.BlockData
	for _, v := range chain.Chain {
		for _, block := range v {
			blockData := format.BlockData{Height: block.Header.Height, Timestamp: block.Header.Timestamp,
				Hash: block.Header.Hash, ParentHash: block.Header.ParentHash, Size: block.Header.Size,
				Mpt: block.Value.Values, Nonce: block.Header.Nonce}
			blockList = append(blockList, blockData)
		}
	}
	str, err := json.Marshal(blockList)
	return string(str), err
}

func (chain *BlockChain) Decode(str string) {
	var blockList []format.BlockData
	if err := json.Unmarshal([]byte(str), &blockList); err != nil {
		panic(err)
	}
	for _, b := range blockList {
		mpt := p1.MerklePatriciaTrie{}
		for k, v := range b.Mpt {
			mpt.Insert(k, v)
		}
		block := Block{header{Height: b.Height, Timestamp: b.Timestamp, Hash: b.Hash,
			ParentHash: b.ParentHash, Size: b.Size, Nonce: b.Nonce}, mpt}
		chain.Insert(block)
	}
}

func (chain *BlockChain) GetLatestBlocks() ([]Block, bool) {
	return chain.Get(chain.Length)
}

func (chain *BlockChain) GetParentBlock(block Block) (Block, bool) {
	blocks, ok := chain.Get(block.Header.Height - 1)
	if !ok {
		return Block{}, ok
	}
	for _, b := range blocks {
		if b.Header.Hash == block.Header.ParentHash {
			return b, true
		}
	}
	return Block{}, false
}

func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}
