/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package committer

import (
	pb "github.com/hyperledger/fabric-protos-go/peer"

	"github.com/hyperledger-labs/fabric-smart-client/platform/fabric/driver"
)

func (c *committer) handleConfig(block driver.Block, fBlock *pb.FilteredBlock, transactions []*pb.FilteredTransaction, i int, event *TxEvent) {
	if block == nil {
		// Fetch it
		ledger, err := c.network.Ledger(c.channel)
		if err != nil {
			logger.Panicf("cannot get ledger [%s]", err)
		}
		block, err = ledger.GetBlockByNumber(fBlock.Number)
		if err != nil {
			logger.Panicf("cannot get filteredBlock [%s]", err)
		}
	}
	tx := transactions[i]

	logger.Debugf("Committing config transaction [%s]", tx.Txid)

	if len(transactions) != 1 {
		logger.Panicf("Config block should contain only one transaction [%s]", tx.Txid)
	}

	switch tx.TxValidationCode {
	case pb.TxValidationCode_VALID:
		committer, err := c.network.Committer(c.channel)
		if err != nil {
			logger.Panicf("Cannot get Committer [%s]", err)
		}

		if err := committer.CommitConfig(fBlock.Number, block.DataAt(i)); err != nil {
			logger.Panicf("Cannot commit config envelope [%s]", err)
		}
	}
}