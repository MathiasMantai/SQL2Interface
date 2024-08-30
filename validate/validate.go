package valdidate

import (
	"errors"
	"strings"
)

func ValidateCreateStatement(statement string) error {
	statement = strings.ToUpper(statement)

	if !strings.Contains(statement, "CREATE") || !strings.Contains(statement, "TABLE") {
		return errors.New("invalid CREATE statement found")
	}

	return nil
}
