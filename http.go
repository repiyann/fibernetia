package fibernetia

import (
	"strings"

	"github.com/valyala/fasthttp"
)

const (
	headerInertia                 = "X-Inertia"
	headerInertiaLocation         = "X-Inertia-Location"
	headerInertiaPartialData      = "X-Inertia-Partial-Data"
	headerInertiaPartialExcept    = "X-Inertia-Partial-Except"
	headerInertiaPartialComponent = "X-Inertia-Partial-Component"
	headerInertiaVersion          = "X-Inertia-Version"
	headerInertiaReset            = "X-Inertia-Reset"
	headerVary                    = "Vary"
	headerContentType             = "Content-Type"
)

// IsInertiaRequest returns true if the request is an Inertia request.
func IsInertiaRequest(ctx *fasthttp.RequestCtx) bool {
	return len(ctx.Request.Header.Peek(headerInertia)) > 0
}

func setInertiaInResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set(headerInertia, "true")
}

func deleteInertiaInResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Del(headerInertia)
}

func setInertiaVaryInResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set(headerVary, headerInertia)
}

func deleteVaryInResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Del(headerVary)
}

func setInertiaLocationInResponse(ctx *fasthttp.RequestCtx, url string) {
	ctx.Response.Header.Set(headerInertiaLocation, url)
}

func setResponseStatus(ctx *fasthttp.RequestCtx, status int) {
	ctx.Response.SetStatusCode(status)
}

func onlyFromRequest(ctx *fasthttp.RequestCtx) []string {
	header := string(ctx.Request.Header.Peek(headerInertiaPartialData))
	if header == "" {
		return nil
	}
	return strings.Split(header, ",")
}

func exceptFromRequest(ctx *fasthttp.RequestCtx) []string {
	header := string(ctx.Request.Header.Peek(headerInertiaPartialExcept))
	if header == "" {
		return nil
	}
	return strings.Split(header, ",")
}

func resetFromRequest(ctx *fasthttp.RequestCtx) []string {
	header := string(ctx.Request.Header.Peek(headerInertiaReset))
	if header == "" {
		return nil
	}
	return strings.Split(header, ",")
}

func partialComponentFromRequest(ctx *fasthttp.RequestCtx) string {
	return string(ctx.Request.Header.Peek(headerInertiaPartialComponent))
}

func inertiaVersionFromRequest(ctx *fasthttp.RequestCtx) string {
	return string(ctx.Request.Header.Peek(headerInertiaVersion))
}

func redirectResponse(ctx *fasthttp.RequestCtx, url string, status ...int) {
	ctx.Response.Header.Set("Location", url)
	setResponseStatus(ctx, firstOr[int](status, fasthttp.StatusFound))
}

func setJSONResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("application/json")
}

func setHTMLResponse(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType("text/html")
}

func isSeeOtherRedirectMethod(method string) bool {
	return method == fasthttp.MethodPut ||
		method == fasthttp.MethodPatch ||
		method == fasthttp.MethodDelete
}

func refererFromRequest(ctx *fasthttp.RequestCtx) string {
	return string(ctx.Request.Header.Referer())
}
