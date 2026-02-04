package i18n

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

type Translations map[string]string

var (
	locales = make(map[string]Translations)
	mu      sync.RWMutex
)

func LoadTranslations(localePath string) error {
	mu.Lock()
	defer mu.Unlock()

	entries, err := os.ReadDir(localePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			locale := entry.Name()
			filePath := filepath.Join(localePath, locale, "relationships.yaml")
			
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue 
			}

			var config struct {
				Relationships Translations `yaml:"RELATIONSHIPS"`
			}
			
			if err := yaml.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("failed to parse %s: %w", filePath, err)
			}

			locales[locale] = config.Relationships
		}
	}

	return nil
}

func Translate(locale, key string) string {
	mu.RLock()
	defer mu.RUnlock()

	if trans, ok := locales[locale]; ok {
		if val, ok := trans[key]; ok {
			return val
		}
	}
	
	if locale != "en" {
		if trans, ok := locales["en"]; ok {
			if val, ok := trans[key]; ok {
				return val
			}
		}
	}

	return key 
}
