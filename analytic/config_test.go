package analytic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfLogLevelsMappedRight(t *testing.T) {
	testMap := configLogLevelMapping()
	assert.Equal(t, ErrorFlag, testMap["FATAL"], "It should result in ERROR flag")
	assert.Equal(t, InfoFlag, testMap["INFO"], "Info should be equal")
	assert.Equal(t, WarningFlag, testMap["WARN"], "Warning flag should be the same")
	assert.Equal(t, DebugFlag, testMap["FINE"], "Debug flag should be produced")
}
