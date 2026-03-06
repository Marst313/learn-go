package utils

import (
	"errors"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

func MapReminderCreateError(err error) (int, string) {

	var pqErr *pq.Error

	var msg string
	var status int = http.StatusBadRequest

	if !errors.As(err, &pqErr) {
		return http.StatusInternalServerError, msg
	}

	constraint := pqErr.Constraint

	switch {
	case strings.Contains(constraint, "category"):
		msg = "Invalid category value available in ('personal','work','health','shopping','other)"

	case strings.Contains(constraint, "priority"):
		msg = "Invalid priority value available in ('low','medium','high')"

	case strings.Contains(constraint, "user_id"):
		msg = "Invalid user"

	case strings.HasSuffix(constraint, "_fkey"):
		msg = "Invalid reference (related data not found)"

	case strings.HasSuffix(constraint, "_check"):
		msg = "Invalid value for one of the fields"

	case strings.HasSuffix(constraint, "_not_null"):
		msg = "Missing required field"

	default:
		msg = "Failed to create reminder"
		status = http.StatusInternalServerError
	}

	return status, msg

}
