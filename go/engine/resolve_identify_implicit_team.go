// Copyright 2017 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// Resolve a set of assertions to an implicit team, optionally creating the team and identifying the members.

package engine

import (
	"fmt"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol/keybase1"
)

// ResolveIdentifyImplicitTeam is an engine.
type ResolveIdentifyImplicitTeam struct {
	libkb.Contextified
	arg    keybase1.ResolveIdentifyImplicitTeamArg
	result keybase1.ResolveIdentifyImplicitTeamRes
}

// NewResolveIdentifyImplicitTeam creates a ResolveIdentifyImplicitTeam engine.
func NewResolveIdentifyImplicitTeam(g *libkb.GlobalContext, arg keybase1.ResolveIdentifyImplicitTeamArg) *ResolveIdentifyImplicitTeam {
	return &ResolveIdentifyImplicitTeam{
		Contextified: libkb.NewContextified(g),
		arg:          arg,
	}
}

// Name is the unique engine name.
func (e *ResolveIdentifyImplicitTeam) Name() string {
	return "ResolveIdentifyImplicitTeam"
}

// GetPrereqs returns the engine prereqs.
func (e *ResolveIdentifyImplicitTeam) Prereqs() Prereqs {
	return Prereqs{}
}

// RequiredUIs returns the required UIs.
func (e *ResolveIdentifyImplicitTeam) RequiredUIs() []libkb.UIKind {
	return []libkb.UIKind{}
}

// SubConsumers returns the other UI consumers for this engine.
func (e *ResolveIdentifyImplicitTeam) SubConsumers() []libkb.UIConsumer {
	return []libkb.UIConsumer{&Identify2WithUID{}}
}

// Run starts the engine.
func (e *ResolveIdentifyImplicitTeam) Run(ctx *Context) (err error) {
	defer e.G().CTrace(ctx.GetNetContext(), "ResolveIdentifyImplicitTeam", func() error { return err })()

	return fmt.Errorf("@@@ nope, nevermind, engines can't use the team package")
}

func (e *ResolveIdentifyImplicitTeam) Result() keybase1.ResolveIdentifyImplicitTeamRes {
	return e.result
}
