package teams

import (
	"strings"

	"github.com/lezzercringe/avito-test-assignment/internal/errorsx"
)

type Team struct {
	Name      string
	MemberIDs []string
}

func New(name string, memberIDs []string) (*Team, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errorsx.ErrTeamName
	}

	s := make(map[string]struct{}, len(memberIDs))
	for _, id := range memberIDs {
		if _, ok := s[id]; ok {
			return nil, errorsx.ErrDuplicateMember
		}
		s[id] = struct{}{}
	}

	return &Team{
		Name:      name,
		MemberIDs: memberIDs,
	}, nil
}
