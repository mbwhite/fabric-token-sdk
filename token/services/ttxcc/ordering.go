/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ttxcc

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/pkg/errors"

	"github.com/hyperledger-labs/fabric-token-sdk/token/services/network"
)

type orderingView struct {
	tx *Transaction
}

// NewOrderingView returns a new instance of the orderingView struct.
// The view does the following:
// 1. It broadcasts the token transaction to the proper Fabric ordering service.
func NewOrderingView(tx *Transaction) *orderingView {
	return &orderingView{tx: tx}
}

// Call execute the view.
// The view does the following:
// 1. It broadcasts the token token transaction to the proper Fabric ordering service.
func (o *orderingView) Call(context view.Context) (interface{}, error) {
	if err := network.GetInstance(context, o.tx.Network(), "").Broadcast(o.tx.Payload.Envelope); err != nil {
		return nil, err
	}
	return nil, nil
}

type orderingAndFinalityView struct {
	tx *Transaction
}

// NewOrderingAndFinalityView returns a new instance of the orderingAndFinalityView struct.
// The view does the following:
// 1. It broadcasts the token transaction to the proper Fabric ordering service.
// 2. It waits for finality of the token transaction by listening to delivery events from one of the
// Fabric peer nodes trusted by the FSC node.
func NewOrderingAndFinalityView(tx *Transaction) *orderingAndFinalityView {
	return &orderingAndFinalityView{tx: tx}
}

// Call executes the view.
// The view does the following:
// 1. It broadcasts the token transaction to the proper Fabric ordering service.
// 2. It waits for finality of the token transaction by listening to delivery events from one of the
// Fabric peer nodes trusted by the FSC node.
func (o *orderingAndFinalityView) Call(context view.Context) (interface{}, error) {
	nw := network.GetInstance(context, o.tx.Network(), o.tx.Channel())
	if nw == nil {
		return nil, errors.Errorf("network [%s] not found", o.tx.Network())
	}
	if err := nw.Broadcast(o.tx.Payload.Envelope); err != nil {
		return nil, err
	}

	return nil, nw.IsFinal(o.tx.ID())
}