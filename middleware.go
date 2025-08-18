package fibernetia

import (
	"bytes"

	"github.com/gofiber/fiber/v2"
)

// Middleware returns Inertia middleware handler for Fiber.
// All handlers that can be handled by Inertia should be wrapped with this.
func (i *Inertia) Middleware(next fiber.Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// Set header Vary to "X-Inertia".
		//
		// https://github.com/inertiajs/inertia-laravel/pull/404
		setInertiaVaryInResponse(ctx)
		setInertiaVersionInResponse(ctx, i.version)

		// Resolve validation errors and clear history from the flash data provider.
		{
			ctx = i.resolveValidationErrors(ctx)
			ctx = i.resolveClearHistory(ctx)
		}

		if !IsInertiaRequest(ctx) {
			return next(ctx)
		}

		// Now we know that this request was made by Inertia.
		//
		// But there is one problem:
		// fiber.Ctx has no methods for getting the response status code and response content.
		// So, we have to create our own response writer wrapper, that will contain that info.
		//
		// It's not critical that we will have a byte buffer, because we
		// know that Inertia response always in JSON format and actually not very big.
		w2 := buildInertiaResponseWrapper(ctx)

		// Now put our response writer wrapper to other handlers.
		err := next(ctx)

		// Determines what to do when the Inertia asset version has changed.
		// By default, we'll initiate a client-side location visit to force an update.
		//
		// https://inertiajs.com/asset-versioning
		if ctx.Method() == fiber.MethodGet && inertiaVersionFromRequest(ctx) != i.version {
			i.Location(ctx, string(ctx.Request().RequestURI()))
			return nil
		}

		// Our response writer wrapper does have all needle data! Yuppy!
		//
		// Don't forget to copy all data to the original
		// response writer before end!
		defer i.copyWrapperResponse(ctx, w2)

		// Determines what to do when an Inertia action returned empty response.
		// By default, we will redirect the user back to where he came from.
		if w2.StatusCode() == fiber.StatusOK && w2.IsEmpty() {
			i.Back(ctx)
		}

		// The PUT, PATCH and DELETE requests cannot have the 302 code status.
		// Let's set the status code to the 303 instead.
		//
		// https://inertiajs.com/redirects#303-response-code
		if w2.StatusCode() == fiber.StatusFound && isSeeOtherRedirectMethod(ctx.Method()) {
			ctx.Status(fiber.StatusSeeOther)
		}

		return err
	}
}

func (i *Inertia) resolveValidationErrors(ctx *fiber.Ctx) *fiber.Ctx {
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

	SetValidationErrors(ctx, validationErrors)
	return ctx
}

func (i *Inertia) resolveClearHistory(ctx *fiber.Ctx) *fiber.Ctx {
	if i.flash == nil {
		return ctx
	}

	clearHistory, err := i.flash.ShouldClearHistory(ctx)
	if err != nil {
		i.logger.Printf("get clear history flag from flash provider error: %s", err)
		return ctx
	}

	if clearHistory {
		ClearHistory(ctx)
	}

	return ctx
}

func (i *Inertia) copyWrapperResponse(ctx *fiber.Ctx, w2 *inertiaResponseWrapper) {
	i.copyWrapperHeaders(ctx, w2)
	i.copyWrapperStatusCode(ctx, w2)
	i.copyWrapperBuffer(ctx, w2)
}

func (i *Inertia) copyWrapperBuffer(ctx *fiber.Ctx, w2 *inertiaResponseWrapper) {
	if err := ctx.SendStream(w2.buf); err != nil {
		i.logger.Printf("cannot copy inertia response buffer: %s", err)
	}
}

func (i *Inertia) copyWrapperStatusCode(ctx *fiber.Ctx, w2 *inertiaResponseWrapper) {
	ctx.Status(w2.statusCode)
}

func (i *Inertia) copyWrapperHeaders(ctx *fiber.Ctx, w2 *inertiaResponseWrapper) {
	for k, v := range w2.header {
		ctx.Set(k, v)
	}
}

type inertiaResponseWrapper struct {
	statusCode int
	buf        *bytes.Buffer
	header     map[string]string
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

func buildInertiaResponseWrapper(ctx *fiber.Ctx) *inertiaResponseWrapper {
	headers := make(map[string]string)
	ctx.Response().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})

	return &inertiaResponseWrapper{
		statusCode: ctx.Response().StatusCode(),
		buf:        bytes.NewBuffer(nil),
		header:     headers,
	}
}
