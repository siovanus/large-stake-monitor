package service

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/ontio/large-stake-monitor/config"
	"github.com/ontio/large-stake-monitor/log"
	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/vm/neovm"
	"os"
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
				continue
			}
			for _, tx := range block.Transactions {
				invokeCode, ok := tx.Payload.(*payload.InvokeCode)
				if !ok {
					continue
				}
				param, err := ParsePayload(invokeCode.Code)
				if err != nil {
					log.Errorf("[Monitor] ParsePayload error:", err)
				}
				for index, pos := range param.PosList {
					if pos > this.config.Limit {
						err := Record(param.Address.ToBase58(), param.PeerPubkeyList[index], pos)
						if err != nil {
							log.Errorf("[Monitor] Record error:", err)
						}
					}
				}
			}
			this.syncHeight++
		}
	}
}

func ParsePayload(code []byte) (*governance.AuthorizeForPeerParam, error) {
	l := len(code)
	param := new(governance.AuthorizeForPeerParam)
	if l > 64 && string(code[l-22:]) == "Ontology.Native.Invoke" && string(code[l-46-18:l-46]) == "unAuthorizeForPeer" {
		executor := neovm.NewExecutor(code)
		err := executor.Execute()
		if err != nil {
			return nil, err
		}

		paramBytes, err := executor.EvalStack.PopAsBytes()
		if err != nil {
			return nil, err
		}

		err = param.Deserialize(bytes.NewBuffer(paramBytes))
		if err != nil {
			return nil, err
		}
	}

	return param, nil
}

func Record(address, pubKey string, pos uint32) error {
	f, err := os.OpenFile("record", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("os.OpenFile error: %s", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	str := fmt.Sprintf("Found large amount unauthorization, address:%s, node public key:%s, value:%d",
		address, pubKey, pos)
	w.WriteString(str)
	w.WriteString("\n")
	w.Flush()
	log.Infof("[Record] %s", str)
	return nil
}
