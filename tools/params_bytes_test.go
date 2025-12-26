// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoolToBytes(t *testing.T) {
	buf := BoolToBytes(true)
	v := BytesToBool(buf)
	assert.Equal(t, true, v)
}

func TestInt32ToBytes(t *testing.T) {
	n := int32(64)
	buf := Int32ToBytes(n)
	v := BytesToInt32(buf)
	assert.Equal(t, n, v)
}

func TestInt64ToBytes(t *testing.T) {
	n := int64(64)
	buf := Int64ToBytes(n)
	v := BytesToInt64(buf)
	assert.Equal(t, n, v)
}

func TestFloat32ToByte(t *testing.T) {
	n := float32(64.043)
	buf := Float32ToBytes(n)
	v := BytesToFloat32(buf)
	assert.Equal(t, n, v)
}

func TestFloat64ToByte(t *testing.T) {
	n := float64(64.043)
	buf := Float64ToBytes(n)
	v := BytesToFloat64(buf)
	assert.Equal(t, n, v)
}
