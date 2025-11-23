package teams_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
	"github.com/lezzercringe/avito-test-assignment/internal/teams"
)

func TestNewTeam(t *testing.T) {
	t.Run("valid team creation", func(t *testing.T) {
		name := "alpha-team"
		memberIDs := []string{"user1", "user2", "user3"}

		team, err := teams.New(name, memberIDs)

		require.NoError(t, err)
		assert.Equal(t, name, team.Name)
		assert.Equal(t, memberIDs, team.MemberIDs)
	})

	t.Run("empty name validation", func(t *testing.T) {
		_, err := teams.New("", []string{"user1"})

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrTeamName, err)
	})

	t.Run("whitespace name validation", func(t *testing.T) {
		_, err := teams.New("   ", []string{"user1"})

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrTeamName, err)
	})

	t.Run("duplicate member validation", func(t *testing.T) {
		memberIDs := []string{"user1", "user2", "user1"}

		_, err := teams.New("test-team", memberIDs)

		assert.Error(t, err)
		assert.Equal(t, errorsx.ErrDuplicateMember, err)
	})
}
