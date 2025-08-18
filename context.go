package fibernetia

import (
	"github.com/gofiber/fiber/v2"
)

// SetTemplateData sets template data to the fiber context.
func SetTemplateData(ctx *fiber.Ctx, templateData TemplateData) *fiber.Ctx {
	ctx.Locals("templateData", templateData)
	return ctx
}

// SetTemplateDatum sets single template data item to the fiber context.
func SetTemplateDatum(ctx *fiber.Ctx, key string, val any) *fiber.Ctx {
	templateData := TemplateDataFromContext(ctx)
	templateData[key] = val

	return SetTemplateData(ctx, templateData)
}

// TemplateDataFromContext returns template data from the fiber context.
func TemplateDataFromContext(ctx *fiber.Ctx) TemplateData {
	val := ctx.Locals("templateData")
	if td, ok := val.(TemplateData); ok {
		return td
	}

	return TemplateData{}
}

// SetProps sets props values to the fiber context.
func SetProps(ctx *fiber.Ctx, props Props) *fiber.Ctx {
	ctx.Locals("props", props)

	return ctx
}

// SetProp sets prop value to the fiber context.
func SetProp(ctx *fiber.Ctx, key string, val any) *fiber.Ctx {
	props := PropsFromContext(ctx)
	props[key] = val

	return SetProps(ctx, props)
}

// PropsFromContext returns props from the fiber context.
func PropsFromContext(ctx *fiber.Ctx) Props {
	val := ctx.Locals("props")
	if p, ok := val.(Props); ok {
		return p
	}

	return Props{}
}

// SetValidationErrors sets validation errors to the fiber context.
func SetValidationErrors(ctx *fiber.Ctx, errors ValidationErrors) *fiber.Ctx {
	ctx.Locals("validationErrors", errors)

	return ctx
}

// AddValidationErrors appends validation errors to the fiber context.
func AddValidationErrors(ctx *fiber.Ctx, errors ValidationErrors) *fiber.Ctx {
	validationErrors := ValidationErrorsFromContext(ctx)
	for key, val := range errors {
		validationErrors[key] = val
	}

	return SetValidationErrors(ctx, validationErrors)
}

// SetValidationError sets validation error to the fiber context.
func SetValidationError(ctx *fiber.Ctx, key string, msg string) *fiber.Ctx {
	validationErrors := ValidationErrorsFromContext(ctx)
	validationErrors[key] = msg

	return SetValidationErrors(ctx, validationErrors)
}

// ValidationErrorsFromContext returns validation errors from the fiber context.
func ValidationErrorsFromContext(ctx *fiber.Ctx) ValidationErrors {
	val := ctx.Locals("validationErrors")
	if ve, ok := val.(ValidationErrors); ok {
		return ve
	}

	return ValidationErrors{}
}

// SetEncryptHistory enables or disables history encryption.
func SetEncryptHistory(ctx *fiber.Ctx, encrypt ...bool) *fiber.Ctx {
	ctx.Locals("encryptHistory", firstOr(encrypt, true))

	return ctx
}

// EncryptHistoryFromContext returns history encryption value from the fiber context.
func EncryptHistoryFromContext(ctx *fiber.Ctx) (bool, bool) {
	val := ctx.Locals("encryptHistory")
	b, ok := val.(bool)

	return b, ok
}

// ClearHistory cleaning history state.
func ClearHistory(ctx *fiber.Ctx) *fiber.Ctx {
	ctx.Locals("clearHistory", true)

	return ctx
}

// ClearHistoryFromContext returns clear history value from the fiber context.
func ClearHistoryFromContext(ctx *fiber.Ctx) bool {
	val := ctx.Locals("clearHistory")
	b, ok := val.(bool)

	return ok && b
}
