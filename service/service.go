package service

import (
	"fmt"
	"io"

	"github.com/ontio/large-stake-monitor/config"
	"github.com/ontio/large-stake-monitor/log"
	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/common"
	common2 "github.com/ontio/ontology/common"
)

type SyncService struct {
	sdk        *sdk.OntologySdk
	syncHeight uint32
	config     *config.Config
}

func NewSyncService(sdk *sdk.OntologySdk) *SyncService {
	syncSvr := &SyncService{
		sdk:    sdk,
		config: config.DefConfig,
	}
	return syncSvr
}

func (this *SyncService) Run() {
	go this.Monitor()
}

func (this *SyncService) Monitor() {
	currentMainChainHeight, err := this.sdk.GetCurrentBlockHeight()
	if err != nil {
		log.Errorf("[Monitor] this.sdk.GetCurrentBlockHeight error:", err)
	}
	this.syncHeight = currentMainChainHeight
	for {
		currentMainChainHeight, err := this.sdk.GetCurrentBlockHeight()
		if err != nil {
			log.Errorf("[Monitor] this.mainSdk.GetCurrentBlockHeight error:", err)
		}
		for i := this.syncHeight; i < currentMainChainHeight; i++ {
			log.Infof("[Monitor] start parse block %d", i)
			//sync key header
			block, err := this.sdk.GetBlockByHeight(i)
			if err != nil {
				log.Errorf("[Monitor] this.mainSdk.GetBlockByHeight error:", err)
			}
			for _, tx := range block.Transactions {
				tx.Payload
			}

			this.syncHeight++
		}
	}
}

func ParsePayload(code []byte) (map[string]interface{}, error) {
	codeHex := common.ToHexString(code)
	l := len(code)
	if l > 44 && string(code[l-22:]) == "Ontology.Native.Invoke" {
		fmt.Println("codeHex:", codeHex)
		if l > 54 && string(code[l-46-8:l-46]) == "transfer" {
			source := common.NewZeroCopySource(code)
			err := ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			from, err := readAddress(source)
			if err != nil {
				return nil, err
			}
			res := make(map[string]interface{})
			res["functionName"] = "transfer"
			res["from"] = from.ToBase58()
			err = ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			to, err := readAddress(source)
			if err != nil {
				return nil, err
			}
			res["to"] = to.ToBase58()
			err = ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			var amount = uint64(0)
			if string(codeHex[source.Pos()*2]) == "5" || string(codeHex[source.Pos()*2]) == "6" {
				//b := common.BigIntFromNeoBytes([]byte{code[source.Pos()]})
				//amount = b.Uint64() - 0x50
				data, eof := source.NextByte()
				if eof {
					return nil, io.ErrUnexpectedEOF
				}
				b := common.BigIntFromNeoBytes([]byte{data})
				amount = b.Uint64() - 0x50
			} else {
				//amount = common.BigIntFromNeoBytes(code[source.Pos()+1 : source.Pos()+1+uint64(code[source.Pos()])]).Uint64()
				amountBytes, _, irregular, eof := source.NextVarBytes()
				if irregular || eof {
					return nil, io.ErrUnexpectedEOF
				}
				amount = common.BigIntFromNeoBytes(amountBytes).Uint64()
			}

			res["amount"] = amount
			if common.ToHexString(common2.ToArrayReverse(code[l-25-20:l-25])) == ONT_CONTRACT_ADDRESS.ToHexString() {
				res["asset"] = "ont"
			} else if common.ToHexString(common2.ToArrayReverse(code[l-25-20:l-25])) == ONG_CONTRACT_ADDRESS.ToHexString() {
				res["asset"] = "ong"
			} else {
				return nil, fmt.Errorf("not ont or ong contractAddress")
			}
			err = ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			//method name
			_, _, irregular, eof := source.NextVarBytes()
			if irregular || eof {
				return nil, io.ErrUnexpectedEOF
			}
			//contract address
			contractAddress, err := readAddress(source)
			if err != nil {
				return nil, err
			}
			res["contractAddress"] = contractAddress
			return res, nil
		} else if l > 58 && string(code[l-46-12:l-46]) == "transferFrom" {
			res := make(map[string]interface{})
			res["functionName"] = "transferFrom"
			source := common.NewZeroCopySource(code)
			err := ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			sender, err := readAddress(source)
			if err != nil {
				return nil, err
			}
			res["sender"] = sender.ToBase58()

			err = ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			from, err := readAddress(source)
			if err != nil {
				return nil, err
			}
			res["from"] = from.ToBase58()
			err = ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			to, err := readAddress(source)
			if err != nil {
				return nil, err
			}
			res["to"] = to.ToBase58()
			err = ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			var amount = uint64(0)
			if string(codeHex[source.Pos()*2]) == "5" || string(codeHex[source.Pos()*2]) == "6" {
				//b := common.BigIntFromNeoBytes([]byte{code[source.Pos()]})
				//amount = b.Uint64() - 0x50
				//read amount
				data, eof := source.NextByte()
				if eof {
					return nil, io.ErrUnexpectedEOF
				}
				b := common.BigIntFromNeoBytes([]byte{data})
				amount = b.Uint64() - 0x50
			} else {
				amountBytes, _, irregular, eof := source.NextVarBytes()
				if irregular || eof {
					return nil, io.ErrUnexpectedEOF
				}
				amount = common.BigIntFromNeoBytes(amountBytes).Uint64()
				//amount = common.BigIntFromNeoBytes(code[source.Pos()+1 : source.Pos()+1+uint64(code[source.Pos()])]).Uint64()
			}
			res["amount"] = amount
			if common.ToHexString(common2.ToArrayReverse(code[l-25-20:l-25])) == ONT_CONTRACT_ADDRESS.ToHexString() {
				res["asset"] = "ont"
			} else if common.ToHexString(common2.ToArrayReverse(code[l-25-20:l-25])) == ONG_CONTRACT_ADDRESS.ToHexString() {
				res["asset"] = "ong"
				res["amount"] = amount
			}
			err = ignoreOpCode(source)
			if err != nil {
				return nil, err
			}
			source.BackUp(1)
			//method name
			_, _, irregular, eof := source.NextVarBytes()
			if irregular || eof {
				return nil, io.ErrUnexpectedEOF
			}
			//contract address
			contractAddress, err := readAddress(source)
			if err != nil {
				return nil, err
			}
			res["contractAddress"] = contractAddress
			return res, nil
		}
	}
	return nil, fmt.Errorf("not native transfer and transferFrom transaction")
}

func readAddress(source *common.ZeroCopySource) (common2.Address, error) {
	senderBytes, _, irregular, eof := source.NextVarBytes()
	if irregular || eof {
		return common.ADDRESS_EMPTY, io.ErrUnexpectedEOF
	}
	sender, err := utils.AddressParseFromBytes(senderBytes)
	if err != nil {
		return common.ADDRESS_EMPTY, err
	}
	return sender, nil
}

func ignoreOpCode(source *common.ZeroCopySource) error {
	opCode := make(map[byte]bool)
	opCode = map[byte]bool{0x00: true, 0xc6: true, 0x6b: true, 0x6a: true, 0xc8: true, 0x6c: true, 0x68: true, 0x67: true,
		0x7c: true, 0x51: true, 0xc1: true}
	s := source.Size()
	for {
		if source.Pos() >= s {
			return nil
		}
		by, eof := source.NextByte()
		if eof {
			return io.EOF
		}
		if opCode[by] {
			continue
		} else {
			return nil
		}
	}
}
