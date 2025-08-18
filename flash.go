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

	// Get existing flash map or create new
	flashMap := map[string]any{}
	if raw := sess.Get("flash"); raw != nil {
		_ = json.Unmarshal(raw.([]byte), &flashMap)
	}

	// Set/overwrite the key
	flashMap[key] = val

	// Save back to session
	data, err := json.Marshal(flashMap)
	if err != nil {
		return err
	}
	sess.Set("flash", data)
	return sess.Save()
}

func (p *SessionFlashProvider) Get(c *fiber.Ctx, key string) (any, error) {
	sess, err := p.store.Get(c)
	if err != nil {
		return nil, err
	}

	raw := sess.Get("flash")
	if raw == nil {
		return nil, nil
	}

	flashMap := map[string]any{}
	if err := json.Unmarshal(raw.([]byte), &flashMap); err != nil {
		return nil, err
	}

	val := flashMap[key]

	delete(flashMap, key)
	data, _ := json.Marshal(flashMap)
	sess.Set("flash", data)

	return val, nil
}

func (p *SessionFlashProvider) GetAll(c *fiber.Ctx) (map[string]any, error) {
	sess, err := p.store.Get(c)
	if err != nil {
		return nil, err
	}
	raw := sess.Get("flash")
	if raw == nil {
		return nil, nil
	}
	flashMap := map[string]any{}
	if err := json.Unmarshal(raw.([]byte), &flashMap); err != nil {
		return nil, err
	}

	sess.Delete("flash")
	_ = sess.Save()
	return flashMap, nil
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
