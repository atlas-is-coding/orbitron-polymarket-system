package i18n

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
)

// Manager holds all loaded locales and the active one.
type Manager struct {
	locales map[string]*Locale
	current atomic.Pointer[Locale]
}

var global Manager

func init() {
	global.locales = make(map[string]*Locale)
	for _, lang := range []string{"en", "ru", "zh", "ja", "ko"} {
		data, err := localeFS.ReadFile("locales/" + lang + ".json")
		if err != nil {
			panic(fmt.Sprintf("i18n: missing locale %s: %v", lang, err))
		}
		var loc Locale
		if err := json.Unmarshal(data, &loc); err != nil {
			panic(fmt.Sprintf("i18n: parse locale %s: %v", lang, err))
		}
		global.locales[lang] = &loc
	}
	// Default to English
	global.current.Store(global.locales["en"])
}

// T returns the current active Locale. Thread-safe, no locks.
func T() *Locale {
	return global.current.Load()
}

// SetLanguage switches the active language. Falls back to "en" if unknown.
func SetLanguage(lang string) {
	if loc, ok := global.locales[lang]; ok {
		global.current.Store(loc)
		return
	}
	global.current.Store(global.locales["en"])
}

// Available returns the list of supported language codes.
func Available() []string {
	return []string{"en", "ru", "zh", "ja", "ko"}
}
