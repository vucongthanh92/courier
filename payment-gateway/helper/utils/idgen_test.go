package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSePayPaymentCode(t *testing.T) {
	code, err := NewSePayPaymentCode("COUR", 8)

	require.NoError(t, err)
	require.Regexp(t, regexp.MustCompile(`^COUR[A-Za-z0-9]{8}$`), code)
}
