/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package manager

import (
	"context"
	"runtime/debug"

	"github.com/pkg/errors"

	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
)

type childContext struct {
	ParentContext view.Context

	session            view.Session
	initiator          view.View
	errorCallbackFuncs []func()
}

func (w *childContext) GetService(v interface{}) (interface{}, error) {
	return w.ParentContext.GetService(v)
}

func (w *childContext) ID() string {
	return w.ParentContext.ID()
}

func (w *childContext) Me() view.Identity {
	return w.ParentContext.Me()
}

func (w *childContext) IsMe(id view.Identity) bool {
	return w.ParentContext.IsMe(id)
}

func (w *childContext) GetSession(caller view.View, party view.Identity) (view.Session, error) {
	return w.ParentContext.GetSession(caller, party)
}

func (w *childContext) GetSessionByID(id string, party view.Identity) (view.Session, error) {
	return w.ParentContext.GetSessionByID(id, party)
}

func (w *childContext) Context() context.Context {
	return w.ParentContext.Context()
}

func (w *childContext) Session() view.Session {
	if w.session == nil {
		return w.ParentContext.Session()
	}
	return w.session
}

func (w *childContext) Initiator() view.View {
	if w.initiator == nil {
		return w.ParentContext.Initiator()
	}
	return w.initiator
}

func (w *childContext) OnError(f func()) {
	w.errorCallbackFuncs = append(w.errorCallbackFuncs, f)
}

func (w *childContext) RunView(v view.View, opts ...view.RunViewOption) (res interface{}, err error) {
	options, err := view.CompileRunViewOptions(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed compiling options")
	}
	var initiator view.View
	if options.AsInitiator {
		initiator = v
	}

	childContext := &childContext{
		ParentContext: w,
		session:       options.Session,
		initiator:     initiator,
	}
	defer func() {
		if r := recover(); r != nil {
			childContext.cleanup()
			res = nil

			logger.Warningf("caught panic while running flow with [%v][%s]", r, debug.Stack())

			switch e := r.(type) {
			case error:
				err = errors.WithMessage(e, "caught panic")
			case string:
				err = errors.Errorf(e)
			default:
				err = errors.Errorf("caught panic [%v]", e)
			}
		}
	}()
	if v == nil && options.Call == nil {
		return nil, errors.Errorf("no view passed")
	}
	if options.Call != nil {
		res, err = options.Call(childContext)
	} else {
		res, err = v.Call(childContext)
	}
	return res, err
}

func (w *childContext) cleanup() {
	for _, callbackFunc := range w.errorCallbackFuncs {
		w.safeInvoke(callbackFunc)
	}
}

func (w *childContext) safeInvoke(f func()) {
	defer func() {
		if r := recover(); r != nil {
			logger.Debugf("function [%s] panicked [%s]", f, r)
		}
	}()
	f()
}