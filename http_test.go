package fibernetia

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestIsInertiaRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header map[string]string
		want   bool
	}{
		{
			"positive",
			map[string]string{"X-Inertia": "foo"},
			true,
		},
		{
			"negative",
			map[string]string{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				for k, v := range tt.header {
					c.Request().Header.Set(k, v)
				}

				got := IsInertiaRequest(c)
				if got != tt.want {
					t.Fatalf("IsInertiaRequest()=%t, want=%t", got, tt.want)
				}

				return nil
			})

		})
	}
}
