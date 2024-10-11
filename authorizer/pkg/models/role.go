package models

import (
	"database/sql/driver"
	"fmt"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser       = "user"
)

func (r *Role) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("type %T is not a string", value)
	}

	switch str {
	case string(RoleAdmin):
		*r = RoleAdmin
	case string(RoleUser):
		*r = RoleUser
	default:
		return fmt.Errorf("role %s is not defined", str)
	}

	return nil
}

func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}
