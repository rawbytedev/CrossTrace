package context

import (
	"crosstrace/internal/configs"
)

// Context store most of configs
/*
We have the main cfg that can be used to applied settings to other fields using
ctx.Share() but we also allow users to configure it on their own if it's needed
each field can be set with different configs and doesn't necessary needs to align with each other
*/
type Context struct {
	cfg     configs.Configs // usually only this needs to be set
	journal configs.JournalConfig
	anchor  configs.AnchorConfig
	batcher configs.BatcherConfig
	minting configs.MintingConfig
}

func NewContext(cfg configs.Configs) *Context {
	return &Context{cfg: cfg}
}

// uses cfg and set the other configs
func (ctx *Context) Share() {
	ctx.journal = ctx.cfg.Journal
	ctx.anchor = ctx.cfg.Anchor
	ctx.batcher = ctx.cfg.Batcher
	ctx.minting = ctx.cfg.Minting
}
