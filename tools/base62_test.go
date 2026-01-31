// Copyright 2014 loolgame Author. All Rights Reserved.
//
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

func TestBase62ToInt(t *testing.T) {
	i := FromBase62("LLqbOL1")
	assert.Equal(t, int64(2713056317011), i)

	i1 := FromBase62("eg")
	assert.Equal(t, int64(884), i1)

	i2 := FromBase62("2cq")
	assert.Equal(t, int64(8458), i2)

	i3 := FromBase62("mim3")
	assert.Equal(t, int64(5313775), i3)
}

func TestIntToBase62(t *testing.T) {
	b := ToBase62(2713056317011)
	assert.Equal(t, "LLqbOL1", b)

	b1 := ToBase62(884)
	assert.Equal(t, "eg", b1)

	b2 := ToBase62(8458)
	assert.Equal(t, "2cq", b2)

	b3 := ToBase62(5313775)
	assert.Equal(t, "mim3", b3)
}
