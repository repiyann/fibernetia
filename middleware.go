package fibernetia

import (
	"bytes"

	"github.com/valyala/fasthttp"
)

// Middleware returns Inertia middleware handler.
// All handlers that can be handled by Inertia should be wrapped with this.
func (i *Inertia) Middleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Set header Vary to "X-Inertia".
		setInertiaVaryInResponse(ctx)

		// Resolve validation errors and clear history from the flash data provider.
		{
			ctx = i.resolveValidationErrors(ctx)
			ctx = i.resolveClearHistory(ctx)
		}

		if !IsInertiaRequest(ctx) {
			next(ctx)
			return
		}

		// Wrap response so we can capture status & body.
		w2 := buildInertiaResponseWrapper(&ctx.Response)

		// Call the next handler with original ctx.
		next(ctx)

		// If Inertia version changed, force client-side reload.
		if string(ctx.Method()) == fasthttp.MethodGet && inertiaVersionFromRequest(ctx) != i.version {
			i.Location(ctx, string(ctx.URI().RequestURI()))
			return
		}

		// Copy buffered response back before finishing.
		defer i.copyWrapperResponse(ctx, w2)

		// Handle empty response (redirect back).
		if w2.StatusCode() == fasthttp.StatusOK && w2.IsEmpty() {
			i.Back(ctx)
		}

		// For PUT/PATCH/DELETE → force 303 instead of 302.
		if w2.StatusCode() == fasthttp.StatusFound && isSeeOtherRedirectMethod(string(ctx.Method())) {
			setResponseStatus(ctx, fasthttp.StatusSeeOther)
		}
	}
}

func (i *Inertia) resolveValidationErrors(ctx *fasthttp.RequestCtx) *fasthttp.RequestCtx {
	if i.flash == nil {
		return ctx
	}

	val, err := i.flash.Get(ctx, "errors")
	if err != nil {
		i.logger.Printf("get validation errors from flash provider error: %s", err)
		return ctx
	}

	validationErrors, ok := val.(ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return ctx
	}

	result := SetValidationErrors(ctx, validationErrors)
	if newCtx, ok := result.(*fasthttp.RequestCtx); ok {
		return newCtx
	}

	return ctx
}

func (i *Inertia) resolveClearHistory(ctx *fasthttp.RequestCtx) *fasthttp.RequestCtx {
	if i.flash == nil {
		return ctx
	}

	clearHistory, err := i.flash.ShouldClearHistory(ctx)
	if err != nil {
		i.logger.Printf("get clear history flag from flash provider error: %s", err)
		return ctx
	}

	if clearHistory {
		if newCtx, ok := ClearHistory(ctx).(*fasthttp.RequestCtx); ok {
			ctx = newCtx
		}
	}

	return ctx
}

func (i *Inertia) copyWrapperResponse(dst *fasthttp.RequestCtx, src *inertiaResponseWrapper) {
	i.copyWrapperHeaders(dst, src)
	i.copyWrapperStatusCode(dst, src)
	i.copyWrapperBuffer(dst, src)
}

func (i *Inertia) copyWrapperBuffer(dst *fasthttp.RequestCtx, src *inertiaResponseWrapper) {
	if _, err := dst.Write(src.buf.Bytes()); err != nil {
		i.logger.Printf("cannot copy inertia response buffer: %s", err)
	}
}

func (i *Inertia) copyWrapperStatusCode(dst *fasthttp.RequestCtx, src *inertiaResponseWrapper) {
	dst.SetStatusCode(src.statusCode)
}

func (i *Inertia) copyWrapperHeaders(dst *fasthttp.RequestCtx, src *inertiaResponseWrapper) {
	src.header.VisitAll(func(k, v []byte) {
		dst.Response.Header.DelBytes(k)
		dst.Response.Header.AddBytesKV(k, v)
	})
}

// inertiaResponseWrapper wraps fasthttp.RequestCtx to buffer output.
type inertiaResponseWrapper struct {
	statusCode int
	buf        *bytes.Buffer
	header     *fasthttp.ResponseHeader
}

func (w *inertiaResponseWrapper) StatusCode() int {
	return w.statusCode
}

func (w *inertiaResponseWrapper) IsEmpty() bool {
	return w.buf.Len() == 0
}

func (w *inertiaResponseWrapper) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *inertiaResponseWrapper) WriteHeader(code int) {
	w.statusCode = code
}

func buildInertiaResponseWrapper(resp *fasthttp.Response) *inertiaResponseWrapper {
	return &inertiaResponseWrapper{
		statusCode: resp.StatusCode(),
		buf:        bytes.NewBuffer(nil),
		header:     &resp.Header,
	}
}
