package fibernetia

import (
	"context"

	"github.com/fasthttp/session"
	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

// SessionFlashProvider implements FlashProvider using fasthttp/session.
type SessionFlashProvider struct {
	store *session.Session
}

func NewSessionFlashProvider(store *session.Session) *SessionFlashProvider {
	return &SessionFlashProvider{store: store}
}

func (p *SessionFlashProvider) Flash(ctx context.Context, key string, val any) error {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	sess, err := p.store.Get(fctx)
	if err != nil {
		return err
	}
	sess.Set("flash_"+key, data)
	return nil
}

func (p *SessionFlashProvider) Get(ctx context.Context, key string) (any, error) {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return nil, nil
	}
	sess, err := p.store.Get(fctx)
	if err != nil {
		return nil, err
	}
	val := sess.Get("flash_" + key)
	if val == nil {
		return nil, nil
	}
	var v any
	err = json.Unmarshal(val.([]byte), &v)
	if err != nil {
		return nil, err
	}
	sess.Delete("flash_" + key) // delete after read
	return v, nil
}

func (p *SessionFlashProvider) FlashClearHistory(ctx context.Context) error {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return nil
	}
	sess, err := p.store.Get(fctx)
	if err != nil {
		return err
	}
	sess.Set("flash_clear_history", true)
	return nil
}

func (p *SessionFlashProvider) ShouldClearHistory(ctx context.Context) (bool, error) {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return false, nil
	}
	sess, err := p.store.Get(fctx)
	if err != nil {
		return false, err
	}
	val := sess.Get("flash_clear_history")
	if val == nil {
		return false, nil
	}
	sess.Delete("flash_clear_history") // delete after read
	return true, nil
}
