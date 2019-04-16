package data

import (
	"../../p1"
	"../../p2"
	"sync"
)

type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Get(height)
}

func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	blocks, ok := sbc.bc.Get(height)
	sbc.mux.Unlock()
	if ok {
		for _, block := range blocks {
			if block.Header.Hash == hash {
				return block, true
			}
		}
	}
	return p2.Block{}, false
}

func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	blocks, ok := sbc.bc.Get(insertBlock.Header.Height - 1)
	sbc.mux.Unlock()
	if ok {
		for _, block := range blocks {
			if block.Header.Hash == insertBlock.Header.ParentHash {
				return true
			}
		}
	}
	return false
}

func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	sbc.bc.Decode(blockChainJson)
	sbc.mux.Unlock()
}

func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Encode()
}

func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GenBlock(mpt)
}

func (sbc *SyncBlockChain) GetLatestBlocks() ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}

func (sbc *SyncBlockChain) Show() string {
	return sbc.bc.Show()
}

func (sbc *SyncBlockChain) GetLength() int32 {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Length
}
