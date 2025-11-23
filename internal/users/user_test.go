package users_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/users"
)

func TestNewUser(t *testing.T) {
	t.Run("valid user creation", func(t *testing.T) {
		id := "user123"
		name := "John Doe"
		active := true

		user, err := users.New(id, name, active)

		require.NoError(t, err)
		assert.Equal(t, id, user.ID)
		assert.Equal(t, name, user.Name)
		assert.Equal(t, active, user.Active)
	})

	t.Run("empty name validation", func(t *testing.T) {
		_, err := users.New("user123", "", true)

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrUserName, err)
	})

	t.Run("whitespace name validation", func(t *testing.T) {
		_, err := users.New("user123", "   ", true)

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrUserName, err)
	})

	t.Run("empty id validation", func(t *testing.T) {
		_, err := users.New("", "John Doe", true)

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrUserID, err)
	})

	t.Run("whitespace id validation", func(t *testing.T) {
		_, err := users.New("   ", "John Doe", true)

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrUserID, err)
	})
}
