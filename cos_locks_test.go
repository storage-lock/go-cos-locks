package cos_locks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalCosLockFactory_InitiallyInitialized(t *testing.T) {
	assert.NotNil(t, cosClientBeanFactory)
}
