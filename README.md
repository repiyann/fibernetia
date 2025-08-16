
# go-fibernetia

Lightweight Inertia-like adapter for fasthttp (and compatible with Fiber) — think Gonertia but for Fiber / fasthttp-based apps.

This package provides server-side helpers to render Inertia-style pages (JSON responses for XHR/Inertia requests and full HTML for normal requests) using fasthttp. It supports shared props, deferred props, SSR passthrough, flashing validation errors, and convenient context helpers.

Note: the codebase uses fasthttp types directly. It is compatible with Fiber (which is built on fasthttp); in Fiber apps you can pass the underlying fasthttp request context into this library.

## Features
- Render Inertia JSON responses for X-Inertia requests and HTML for normal requests
- Shared props, template data, and template funcs
- Deferred props and mergeable props
- Optional/Lazy props and functions that resolve with or without context
- Redirect helpers (Location, Redirect, Back) that behave correctly for Inertia requests
- SSR passthrough: POST rendered page JSON to an SSR endpoint and embed the returned HTML

## Install

This repository uses Go modules. From your module use:

```bash
go get go-fibernetia
```

Then import in your code:

```go
import "go-fibernetia"
```

(If you host the module under a VCS path, replace the module path accordingly.)

## Quick start (fasthttp)

The minimal flow is to create an Inertia instance with a root template and call `Render` from your request handler.

```go
// create inertia (root template contains the HTML shell with a placeholder container)
i, err := gonertia.New(rootTemplateHTML, gonertia.WithVersion("v1"))
if err != nil {
	log.Fatalf("init inertia: %v", err)
}

// a fasthttp handler
func handler(ctx *fasthttp.RequestCtx) {
	props := gonertia.Props{
		"user": map[string]any{"id": 1, "name": "Alice"},
	}
	if err := i.Render(ctx, "Dashboard", props); err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}

// start server
// fasthttp.ListenAndServe(":8080", handler)
```

## Using with Fiber

Fiber is built on top of fasthttp. This library operates on `*fasthttp.RequestCtx`, so you can forward the underlying fasthttp context from a Fiber handler to Inertia. The exact method to get the underlying context depends on the Fiber version; the important point is to pass the fasthttp RequestCtx to Inertia's methods.

Example (conceptual):

```go
// in a Fiber handler you can forward the underlying fasthttp context
// fctx *fiber.Ctx
// inertia.Render(fctx.RequestCtx(), "MyComponent", props)
```

If you prefer, wrap the call in a small adapter to keep handlers idiomatic.

## API overview

- New(rootTemplateHTML string, opts ...Option) (*Inertia, error)
- NewFromFile(path string, opts ...Option) (*Inertia, error)
- NewFromFileFS(fs.FS, path string, opts ...Option) (*Inertia, error)
- Render(ctx *fasthttp.RequestCtx, component string, props ...Props) error
- Location/Redirect/Back helpers for redirects
- ShareProp/SharedProps/ShareTemplateData/ShareTemplateFunc
- WithVersion, WithSSR, WithContainerID, WithJSONMarshaller, WithLogger, WithFlashProvider, WithEncryptHistory

Types and helpers:

- Props: map[string]any — data passed to the client component
- Optional, Defer, Merge helpers for lazy/deferred/mergeable props
- ValidationErrors and flash provider interface for server-side validation
- Context helpers in `context.go` to set props, template data, validation errors and history behavior

## SSR

If you enable SSR with `WithSSR(url)` the library will POST the serialized page JSON to the configured SSR endpoint (default: http://127.0.0.1:13714/render) and embed the returned HTML into the root template. If SSR fails, it falls back to embedding the JSON container and logs the error via the configured logger.

## Example: advanced props

```go
// deferred prop that will be resolved after initial render
props := gonertia.Props{
	"user": gonertia.Always(map[string]any{"id": 1}),
	"stats": gonertia.Defer(func() any {
		// expensive computation or DB query
		return map[string]int{"visits": 100}
	}),
}

_ = i.Render(ctx, "Dashboard", props)
```

## License

This project is provided under the MIT License. See the LICENSE file for details.

## Notes

- The module in this repository is named `go-fibernetia` (see `go.mod`). If you publish the module to a VCS hosting service, update the import path accordingly.
- The README intentionally focuses on usage with fasthttp and the compatibility story for Fiber; see the code comments for additional API details and lower-level helpers.
