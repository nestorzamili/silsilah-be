package unit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"silsilah-keluarga/internal/pkg/i18n"
)

func TestI18nLoading(t *testing.T) {
	// Path to locales relative to this test file
	// tests/unit -> ../../locales
	localePath := filepath.Join("..", "..", "locales")
	
	err := i18n.LoadTranslations(localePath)
	assert.NoError(t, err, "Should load translations without error")

	// Test ID keys
	assert.Equal(t, "Ayah", i18n.Translate("id", "FATHER"))
	assert.Equal(t, "Ibu", i18n.Translate("id", "MOTHER"))
	assert.Equal(t, "Paman", i18n.Translate("id", "UNCLE"))
	
	// Test EN keys (Newly added)
	assert.Equal(t, "Father", i18n.Translate("en", "FATHER"))
	assert.Equal(t, "Mother", i18n.Translate("en", "MOTHER"))
	assert.Equal(t, "Uncle", i18n.Translate("en", "UNCLE"))
	assert.Equal(t, "Granddaughter", i18n.Translate("en", "GRANDDAUGHTER"))
	
	// Test Fallback
	// Assuming a key that doesn't exist in ID but might in EN (though we made them symmetric)
	// Let's test non-existent key returns key
	assert.Equal(t, "NON_EXISTENT_KEY", i18n.Translate("id", "NON_EXISTENT_KEY"))
}
