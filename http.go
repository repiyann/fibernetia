package fibernetia

import (
	"strings"

	"github.com/gofiber/fiber/v2"
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
func IsInertiaRequest(ctx *fiber.Ctx) bool {
	return len(ctx.Request().Header.Peek(headerInertia)) > 0
}

func setInertiaInResponse(ctx *fiber.Ctx) {
	ctx.Response().Header.Set(headerInertia, "true")
}

func deleteInertiaInResponse(ctx *fiber.Ctx) {
	ctx.Response().Header.Del(headerInertia)
}

func setInertiaVaryInResponse(ctx *fiber.Ctx) {
	ctx.Response().Header.Set(headerVary, headerInertia)
}

func deleteVaryInResponse(ctx *fiber.Ctx) {
	ctx.Response().Header.Del(headerVary)
}

func setInertiaVersionInResponse(ctx *fiber.Ctx, version string) {
	ctx.Response().Header.Set(headerInertiaVersion, version)
}

func setInertiaLocationInResponse(ctx *fiber.Ctx, url string) {
	ctx.Response().Header.Set(headerInertiaLocation, url)
}

func setResponseStatus(ctx *fiber.Ctx, status int) {
	ctx.Response().SetStatusCode(status)
}

func onlyFromRequest(ctx *fiber.Ctx) []string {
	header := string(ctx.Request().Header.Peek(headerInertiaPartialData))
	if header == "" {
		return nil
	}

	return strings.Split(header, ",")
}

func exceptFromRequest(ctx *fiber.Ctx) []string {
	header := string(ctx.Request().Header.Peek(headerInertiaPartialExcept))
	if header == "" {
		return nil
	}

	return strings.Split(header, ",")
}

func resetFromRequest(ctx *fiber.Ctx) []string {
	header := string(ctx.Request().Header.Peek(headerInertiaReset))
	if header == "" {
		return nil
	}

	return strings.Split(header, ",")
}

func partialComponentFromRequest(ctx *fiber.Ctx) string {
	return string(ctx.Request().Header.Peek(headerInertiaPartialComponent))
}

func inertiaVersionFromRequest(ctx *fiber.Ctx) string {
	return string(ctx.Request().Header.Peek(headerInertiaVersion))
}

func redirectResponse(ctx *fiber.Ctx, url string, status ...int) {
	ctx.Response().Header.Set("Location", url)
	setResponseStatus(ctx, firstOr(status, fiber.StatusFound))
}

func setJSONResponse(ctx *fiber.Ctx) {
	ctx.Response().Header.Set(headerContentType, "application/json")
}

func setHTMLResponse(ctx *fiber.Ctx) {
	ctx.Response().Header.Set(headerContentType, "text/html")
}

func isSeeOtherRedirectMethod(method string) bool {
	return method == fiber.MethodPut ||
		method == fiber.MethodPatch ||
		method == fiber.MethodDelete
}

func refererFromRequest(ctx *fiber.Ctx) string {
	return string(ctx.Request().Header.Referer())
}
