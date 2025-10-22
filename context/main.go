package context

import (
	"crosstrace/internal/configs"
	"crosstrace/internal/crypto"
	"crosstrace/internal/encoder"
)

// Context store most of configs
/*
We have the main cfg that can be used to applied settings to other fields using
ctx.Share() but we also allow users to configure it on their own if it's needed
each field can be set with different configs and doesn't necessary needs to align with each other
*/
type Context struct {
	Cfg     configs.Configs // usually only this needs to be set
	Journal configs.JournalConfig
	Anchor  configs.AnchorConfig
	Batcher configs.BatcherConfig
	Minting configs.MintingConfig
	Encoder encoder.Encoder
	Hasher  crypto.Hasher
}

func NewContext(cfg configs.Configs) *Context {
	return &Context{Cfg: cfg}
}

// uses cfg and set the other configs
func (ctx *Context) Share() {
	ctx.Journal = ctx.Cfg.Journal
	ctx.Anchor = ctx.Cfg.Anchor
	ctx.Batcher = ctx.Cfg.Batcher
	ctx.Minting = ctx.Cfg.Minting
	ctx.Hasher = crypto.NewHasher(ctx.Journal.HasherName)
	ctx.Encoder = encoder.NewEncoder(ctx.Journal.EncoderName)
}
