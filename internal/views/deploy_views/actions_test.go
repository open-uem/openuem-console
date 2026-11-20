package deploy_views

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAction(t *testing.T) {
	for value, expected := range map[string]string{
		"true":          ActionInstall,
		"false":         ActionUninstall,
		ActionInstall:   ActionInstall,
		ActionUpdate:    ActionUpdate,
		ActionUninstall: ActionUninstall,
		ActionUpdateAll: ActionUpdateAll,
	} {
		action, ok := ParseAction(value)
		assert.True(t, ok, value)
		assert.Equal(t, expected, action, value)
	}

	for _, value := range []string{"", "remove", "TRUE"} {
		_, ok := ParseAction(value)
		assert.False(t, ok, value)
	}
}
