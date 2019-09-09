// Copyright 2019 Cisco Systems, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	of "github.com/cisco-cx/of/lib/v1alpha1"
)

// Confirm that of.Time implements the fmt.Stringer interface.
func TestTime_StringerInterface(t *testing.T) {
	ofTime := of.NewTime(time.Time{})
	var _ fmt.Stringer = ofTime
	assert.Nil(t, nil) // If we get this far, the test passed.
}

// Simple test of Time's implementation of the fmt.Stringer interface.
func TestTime_StringerSimple(t *testing.T) {
	ofTime := of.NewTime(time.Date(2000, time.February, 29, 1, 2, 3, 4, time.UTC))
	assert.Equal(t, "2000-02-29T01:02:03Z", fmt.Sprintf("%s", ofTime))
}

// Simple test of Time's implementation of the Marshaler interface in
// encoding/json.
func TestTime_JSONMarshalerSimple(t *testing.T) {
	assert := assert.New(t) // Prepare to assert multiple times
	input := time.Date(2000, time.February, 29, 1, 2, 3, 4, time.UTC)
	ofTime := of.NewTime(input)
	result, err := ofTime.MarshalJSON()
	assert.Nil(err, "Time.MarshalJSON() returned non-nil error")
	expect := []byte("\"2000-02-29T01:02:03Z\"")
	assert.Equal(expect, result, "Did not obtain expected result.")
}

// Confirm that of.Time implements the Marshaler interface in encoding/json.
func TestTime_JSONMarshalerInterface(t *testing.T) {
	ofTime := of.NewTime(time.Time{})
	var _ json.Marshaler = ofTime
	assert.Nil(t, nil) // If we get this far, the test passed.
}
