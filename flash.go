package fibernetia

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

// CookieFlashProvider implements FlashProvider using cookies.
type CookieFlashProvider struct {
	errorCookieName   string
	historyCookieName string
}

func NewCookieFlashProvider() *CookieFlashProvider {
	return &CookieFlashProvider{
		errorCookieName:   "flash_errors",
		historyCookieName: "flash_clear_history",
	}
}

func (p *CookieFlashProvider) Flash(ctx context.Context, key string, val any) error {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return nil
	}

	data, _ := json.Marshal(val)
	cookie := fasthttp.AcquireCookie()
	cookie.SetKey("flash_" + key)
	cookie.SetValueBytes(data)
	cookie.SetPath("/")
	fctx.Response.Header.SetCookie(cookie)
	return nil
}

func (p *CookieFlashProvider) Get(ctx context.Context, key string) (any, error) {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return nil, nil
	}

	cookie := fctx.Request.Header.Cookie("flash_" + key)
	if len(cookie) == 0 {
		return nil, nil
	}

	var v any
	_ = json.Unmarshal(cookie, &v)

	// delete after read
	fctx.Response.Header.DelCookie("flash_" + key)

	return v, nil
}

func (p *CookieFlashProvider) FlashClearHistory(ctx context.Context) error {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return nil
	}

	cookie := fasthttp.AcquireCookie()
	cookie.SetKey(p.historyCookieName)
	cookie.SetValue("1")
	cookie.SetPath("/")
	fctx.Response.Header.SetCookie(cookie)

	return nil
}

func (p *CookieFlashProvider) ShouldClearHistory(ctx context.Context) (bool, error) {
	fctx, ok := ctx.(*fasthttp.RequestCtx)
	if !ok {
		return false, nil
	}

	cookie := fctx.Request.Header.Cookie(p.historyCookieName)
	if len(cookie) == 0 {
		return false, nil
	}

	// Delete after read
	fctx.Response.Header.DelCookie(p.historyCookieName)

	return true, nil
}
