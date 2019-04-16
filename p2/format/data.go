package format

type BlockData struct {
	Height     int32             `json:"height"`
	Timestamp  int64             `json:"timeStamp"`
	Hash       string            `json:"hash"`
	ParentHash string            `json:"parentHash"`
	Size       int32             `json:"size"`
	Mpt        map[string]string `json:"mpt"`
	Nonce      string            `json:"nonce"`
}

func (blockData *BlockData) Inital(Height int32, Timestamp int64, Hash string, ParentHash string, Size int32, Mpt map[string]string, Nonce string) {
	blockData.Height = Height
	blockData.Timestamp = Timestamp
	blockData.Hash = Hash
	blockData.ParentHash = ParentHash
	blockData.Size = Size
	blockData.Mpt = Mpt
	blockData.Nonce = Nonce
}
