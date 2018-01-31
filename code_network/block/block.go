package block

import (
	"crypto/sha256"
	"fmt"
)

var GenesisBlock = &Block{
	Index:        0,
	PreviousHash: "0",
	Timestamp:    1465154705,
	Data:         "my genesis block!!",
	Hash:         "816534932c2b7154836da6afc367695e6337db8a921823784c14378abed4f7d7",
}

type Blocks []*Block

func (blocks Blocks) Len() int {
	return len(blocks)
}

func (blocks Blocks) Swap(i, j int) {
	blocks[i], blocks[j] = blocks[j], blocks[i]
}

func (blocks Blocks) Less(i, j int) bool {
	return blocks[i].Index < blocks[j].Index
}

type Block struct {
	Index        int64  `json:"index"`
	PreviousHash string `json:"previousHash"`
	Timestamp    int64  `json:"timestamp"`
	Data         string `json:"data"`
	Hash         string `json:"hash"`
}

func (block *Block) Hhash() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf(
		"%d%s%d%s",
		block.Index, block.PreviousHash, block.Timestamp, block.Data,
	))))
}

func (block *Block) IsValidHash() bool {
	return block.Hash == block.Hhash()
}

func IsValidBlock(block, prevBlock *Block) bool {
	return block.Index == prevBlock.Index+1 &&
		block.PreviousHash == prevBlock.Hash &&
		block.IsValidHash()
}