package fibernetia

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// SessionFlashProvider implements FlashProvider using github.com/gofiber/session.
type SessionFlashProvider struct {
	store *session.Store
}

func NewSessionFlashProvider(store *session.Store) *SessionFlashProvider {
	return &SessionFlashProvider{store: store}
}

func (p *SessionFlashProvider) Flash(c *fiber.Ctx, key string, val any) error {
	sess, err := p.store.Get(c)
	if err != nil {
		return err
	}
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	sess.Set("flash_"+key, data)
	return nil
}

func (p *SessionFlashProvider) Get(c *fiber.Ctx, key string) (any, error) {
	sess, err := p.store.Get(c)
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

func (p *SessionFlashProvider) FlashClearHistory(c *fiber.Ctx) error {
	sess, err := p.store.Get(c)
	if err != nil {
		return err
	}
	sess.Set("flash_clear_history", true)
	return nil
}

func (p *SessionFlashProvider) ShouldClearHistory(c *fiber.Ctx) (bool, error) {
	sess, err := p.store.Get(c)
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
