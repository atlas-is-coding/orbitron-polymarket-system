package i18n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/atlasdev/polytrade-bot/internal/i18n"
)

func TestT_ReturnsNonNil(t *testing.T) {
	loc := i18n.T()
	require.NotNil(t, loc)
}

func TestT_DefaultEnglish(t *testing.T) {
	i18n.SetLanguage("en")
	loc := i18n.T()
	assert.NotEmpty(t, loc.TabOverview)
	assert.NotEmpty(t, loc.TabOrders)
	assert.NotEmpty(t, loc.TabPositions)
}

func TestSetLanguage_AllLocales(t *testing.T) {
	langs := i18n.Available()
	require.Equal(t, []string{"en", "ru", "zh", "ja", "ko"}, langs)
	for _, lang := range langs {
		t.Run(lang, func(t *testing.T) {
			i18n.SetLanguage(lang)
			loc := i18n.T()
			require.NotNil(t, loc)
			assert.NotEmpty(t, loc.TabOverview, "TabOverview empty for lang %s", lang)
			assert.NotEmpty(t, loc.TabOrders, "TabOrders empty for lang %s", lang)
		})
	}
}

func TestSetLanguage_UnknownFallsBackToEnglish(t *testing.T) {
	i18n.SetLanguage("en")
	engTab := i18n.T().TabOverview
	i18n.SetLanguage("xyz-unknown")
	loc := i18n.T()
	require.NotNil(t, loc)
	assert.Equal(t, engTab, loc.TabOverview)
}

func TestSetLanguage_ThreadSafe(t *testing.T) {
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			i18n.SetLanguage("ru")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		i18n.SetLanguage("en")
		_ = i18n.T()
	}
	<-done
}

func TestAvailable_Returns5Languages(t *testing.T) {
	langs := i18n.Available()
	assert.Len(t, langs, 5)
}
