package serverpool

import (
	"reflect"
	"testing"

	"github.com/abhikvarma/load-balancer/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewServerPool(t *testing.T) {
	testCases := []struct {
		name         string
		strategy     utils.LBStrategy
		expectedPool interface{}
	}{
		{
			name:         "ValidStrategyRoundRobin",
			strategy:     utils.RoundRobin,
			expectedPool: &roundRobinServerPool{},
		},
		{
			name:         "ValidStrategyLeastConn",
			strategy:     utils.LeastConnected,
			expectedPool: &lcServerPool{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pool, err := NewServerPool(tc.strategy)

			assert.NoError(t, err)
			assert.NotNil(t, pool)

			expectedPoolType := reflect.TypeOf(tc.expectedPool)
			poolType := reflect.TypeOf(pool)

			assert.Equal(t, expectedPoolType, poolType,
				"expected type: %s, actual type: %s", expectedPoolType, poolType)
		})
	}
}
