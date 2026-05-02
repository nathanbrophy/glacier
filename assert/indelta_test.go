// SPDX-License-Identifier: Apache-2.0

package assert

import (
	"math"
	"testing"
)

// §21.4 F6, E13

func TestInDeltaFloat64(t *testing.T) {
	mt := &mockTB{}
	True(t, InDelta(mt, 1.0001, 1.0, 0.001), "InDelta(1.0001, 1.0, 0.001) = true")
	Equal(t, mt.errorfCalls, 0)
	mt.reset()
	False(t, InDelta(mt, 1.01, 1.0, 0.001), "InDelta(1.01, 1.0, 0.001) = false")
	Equal(t, mt.errorfCalls, 1)
}

func TestInDeltaFloat32(t *testing.T) {
	mt := &mockTB{}
	True(t, InDelta[float32](mt, 1.001, 1.0, 0.01), "InDelta float32")
}

type MyFloat float64

func TestInDeltaCustomFloat(t *testing.T) {
	mt := &mockTB{}
	True(t, InDelta[MyFloat](mt, 1.0001, 1.0, 0.001), "InDelta MyFloat (~float64)")
}

func TestInDeltaNaN(t *testing.T) {
	mt := &mockTB{}
	nan := math.NaN()
	False(t, InDelta(mt, nan, 1.0, 100.0), "InDelta NaN → false")
	Equal(t, mt.errorfCalls, 1)
}

func TestInDeltaNegativeDelta(t *testing.T) {
	mt := &mockTB{}
	False(t, InDelta(mt, 1.0, 1.0, -0.1), "InDelta negative delta → false")
	Equal(t, mt.errorfCalls, 1)
}

func TestInDeltaExact(t *testing.T) {
	mt := &mockTB{}
	True(t, InDelta(mt, 1.0, 1.0, 0.0), "InDelta(1.0, 1.0, 0) = true (exact)")
}
