package fibernetia

import (
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestInertia_SetTemplateData(t *testing.T) {
	t.Parallel()

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		SetTemplateData(ctx, TemplateData{"foo": "bar"})

		got := ctx.Locals("templateData").(TemplateData)
		want := TemplateData{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("TemplateData=%#v, want=%#v", got, want)
		}

		app.ReleaseCtx(ctx)
	})

	t.Run("already filled", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx.Locals("templateData", TemplateData{"baz": "quz", "foo": "quz"})
		SetTemplateData(ctx, TemplateData{"foo": "bar"})

		got := ctx.Locals("templateData").(TemplateData)
		want := TemplateData{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("TemplateData=%#v, want=%#v", got, want)
		}

		app.ReleaseCtx(ctx)
	})
}

func TestInertia_SetTemplateDatum(t *testing.T) {
	t.Parallel()

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		SetTemplateDatum(ctx, "foo", "bar")

		got := ctx.Locals("templateData").(TemplateData)
		want := TemplateData{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("TemplateData=%#v, want=%#v", got, want)
		}

		app.ReleaseCtx(ctx)
	})

	t.Run("already filled", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx.Locals("templateData", TemplateData{"baz": "quz", "foo": "quz"})
		SetTemplateDatum(ctx, "foo", "bar")

		got := ctx.Locals("templateData").(TemplateData)
		want := TemplateData{"foo": "bar", "baz": "quz"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("TemplateData=%#v, want=%#v", got, want)
		}

		app.ReleaseCtx(ctx)
	})
}

func Test_TemplateDataFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctxData any
		want    TemplateData
	}{
		{
			name:    "nil",
			ctxData: nil,
			want:    TemplateData{},
		},
		{
			name:    "empty",
			ctxData: TemplateData{},
			want:    TemplateData{},
		},
		{
			name:    "filled",
			ctxData: TemplateData{"foo": "bar"},
			want:    TemplateData{"foo": "bar"},
		},
		{
			name:    "wrong type",
			ctxData: []string{"foo", "bar"},
			want:    TemplateData{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

			ctx.Locals("templateData", tt.ctxData)

			got := TemplateDataFromContext(ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("TemplateData=%#v, want=%#v", got, tt.want)
			}
		})
	}
}

func TestInertia_SetProps(t *testing.T) {
	t.Parallel()

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx = SetProps(ctx, Props{"foo": "bar"})

		got, ok := ctx.Locals("props").(Props)
		if !ok {
			t.Fatal("props from context is not `Props` type")
		}

		want := Props{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Props=%#v, want=%#v", got, want)
		}
	})

	t.Run("already filled", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx.Locals("props", Props{"baz": "quz", "foo": "quz"})
		ctx = SetProps(ctx, Props{"foo": "bar"})

		got, ok := ctx.Locals("props").(Props)
		if !ok {
			t.Fatal("props from context is not `Props` type")
		}

		want := Props{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Props=%#v, want=%#v", got, want)
		}
	})
}

func TestInertia_SetProp(t *testing.T) {
	t.Parallel()

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx = SetProp(ctx, "foo", "bar")

		got, ok := ctx.Locals("props").(Props)
		if !ok {
			t.Fatal("props from context is not `Props` type")
		}

		want := Props{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("Props=%#v, want=%#v", got, want)
		}
	})

	t.Run("already filled", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx.Locals("props", Props{"baz": "quz", "foo": "quz"})
		ctx = SetProp(ctx, "foo", "bar")

		got, ok := ctx.Locals("props").(Props)
		if !ok {
			t.Fatal("props from context is not `Props` type")
		}

		want := Props{"foo": "bar", "baz": "quz"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("props=%#v, want=%#v", got, want)
		}
	})
}

func Test_PropsFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctxData any
		want    Props
	}{
		{
			name:    "nil",
			ctxData: nil,
			want:    Props{},
		},
		{
			name:    "empty",
			ctxData: Props{},
			want:    Props{},
		},
		{
			name:    "filled",
			ctxData: Props{"foo": "bar"},
			want:    Props{"foo": "bar"},
		},
		{
			name:    "wrong type",
			ctxData: []string{"foo", "bar"},
			want:    Props{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

			ctx.Locals("props", tt.ctxData)

			got := PropsFromContext(ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Props=%#v, want=%#v", got, tt.want)
			}
		})
	}
}

func TestInertia_SetValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx = SetValidationErrors(ctx, ValidationErrors{"foo": "bar"})

		got, ok := ctx.Locals("validationErrors").(ValidationErrors)
		if !ok {
			t.Fatal("validation errors from context is not `ValidationErrors` type")
		}

		want := ValidationErrors{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ValidationErrors=%#v, want=%#v", got, want)
		}
	})

	t.Run("already filled", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx.Locals("validationErrors", ValidationErrors{"baz": "quz", "foo": "quz"})
		ctx = SetValidationErrors(ctx, ValidationErrors{"foo": "bar"})

		got, ok := ctx.Locals("validationErrors").(ValidationErrors)
		if !ok {
			t.Fatal("validation errors from context is not `ValidationErrors` type")
		}

		want := ValidationErrors{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ValidationErrors=%#v, want=%#v", got, want)
		}
	})
}

func TestInertia_AddValidationErrors(t *testing.T) {
	t.Parallel()

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx = AddValidationErrors(ctx, ValidationErrors{"foo": "bar"})

		got, ok := ctx.Locals("validationErrors").(ValidationErrors)
		if !ok {
			t.Fatal("validation errors from context is not `ValidationErrors` type")
		}

		want := ValidationErrors{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ValidationErrors=%#v, want=%#v", got, want)
		}
	})

	t.Run("already filled", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx.Locals("validationErrors", ValidationErrors{"baz": "quz", "foo": "quz"})
		ctx = AddValidationErrors(ctx, ValidationErrors{"foo": "bar"})

		got, ok := ctx.Locals("validationErrors").(ValidationErrors)
		if !ok {
			t.Fatal("validation errors from context is not `ValidationErrors` type")
		}

		want := ValidationErrors{"baz": "quz", "foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ValidationErrors=%#v, want=%#v", got, want)
		}
	})
}

func TestInertia_SetValidationError(t *testing.T) {
	t.Parallel()

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
		ctx = SetValidationError(ctx, "foo", "bar")

		got, ok := ctx.Locals("validationErrors").(ValidationErrors)
		if !ok {
			t.Fatal("validation errors from context is not `ValidationErrors` type")
		}

		want := ValidationErrors{"foo": "bar"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ValidationErrors=%#v, want=%#v", got, want)
		}
	})

	t.Run("already filled", func(t *testing.T) {
		t.Parallel()

		app := fiber.New()
		ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

		ctx.Locals("validationErrors", ValidationErrors{"baz": "quz", "foo": "quz"})
		ctx = SetValidationError(ctx, "foo", "bar")

		got, ok := ctx.Locals("validationErrors").(ValidationErrors)
		if !ok {
			t.Fatal("validation errors from context is not `ValidationErrors` type")
		}

		want := ValidationErrors{"foo": "bar", "baz": "quz"}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ValidationErrors=%#v, want=%#v", got, want)
		}
	})
}

func Test_ValidationErrorsFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctxData any
		want    ValidationErrors
	}{
		{
			name:    "nil",
			ctxData: nil,
			want:    ValidationErrors{},
		},
		{
			name:    "empty",
			ctxData: ValidationErrors{},
			want:    ValidationErrors{},
		},
		{
			name:    "filled",
			ctxData: ValidationErrors{"foo": "bar"},
			want:    ValidationErrors{"foo": "bar"},
		},
		{
			name:    "filled with nested",
			ctxData: ValidationErrors{"foo": ValidationErrors{"abc": "123"}},
			want:    ValidationErrors{"foo": ValidationErrors{"abc": "123"}},
		},
		{
			name:    "wrong type",
			ctxData: []string{"foo", "bar"},
			want:    ValidationErrors{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

			ctx.Locals("validationErrors", tt.ctxData)

			got := ValidationErrorsFromContext(ctx)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ValidationErrors=%#v, want=%#v", got, tt.want)
			}
		})
	}
}

func TestInertia_SetEncryptHistory(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

	ctx = SetEncryptHistory(ctx)

	got, ok := ctx.Locals("encryptHistory").(bool)
	if !ok {
		t.Fatal("encrypt history from context is not `bool` type")
	}

	want := true

	if got != want {
		t.Fatalf("encryptHistory=%t, want=%t", got, want)
	}
}

func Test_EncryptHistoryFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctxData any
		want    bool
	}{
		{
			name:    "nil",
			ctxData: nil,
			want:    false,
		},
		{
			name:    "false",
			ctxData: false,
			want:    false,
		},
		{
			name:    "true",
			ctxData: true,
			want:    true,
		},
		{
			name:    "wrong type",
			ctxData: []string{"foo", "bar"},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
			ctx.Locals("encryptHistory", tt.ctxData)

			got, _ := EncryptHistoryFromContext(ctx)
			if got != tt.want {
				t.Fatalf("encryptHistory=%t, want=%t", got, tt.want)
			}
		})
	}
}

func TestInertia_ClearHistory(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

	ctx = ClearHistory(ctx)

	got, ok := ctx.Locals("clearHistory").(bool)
	if !ok {
		t.Fatal("clear history from context is not `bool` type")
	}

	want := true

	if got != want {
		t.Fatalf("clearHistory=%t, want=%t", got, want)
	}
}

func Test_ClearHistoryFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctxData any
		want    bool
	}{
		{
			name:    "nil",
			ctxData: nil,
			want:    false,
		},
		{
			name:    "false",
			ctxData: false,
			want:    false,
		},
		{
			name:    "true",
			ctxData: true,
			want:    true,
		},
		{
			name:    "wrong type",
			ctxData: []string{"foo", "bar"},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			ctx := app.AcquireCtx(&fasthttp.RequestCtx{})

			ctx.Locals("clearHistory", tt.ctxData)

			got := ClearHistoryFromContext(ctx)
			if got != tt.want {
				t.Fatalf("clearHistory=%t, want=%t", got, tt.want)
			}
		})
	}
}
