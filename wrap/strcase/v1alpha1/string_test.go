// Copyright 2019 Cisco Systems, Inc.
//
// This work incorporates works covered by the following notice:
//
// Copyright 2018 The Prometheus Authors
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
//
// This work incorporates works covered by the following notices:
//
// The MIT License (MIT)
//
// Copyright (c) 2015 Ian Coleman
// Copyright (c) 2018 Ma_124, <github.com/Ma124>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, Subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or Substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package v1alpha1_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	of "github.com/cisco-cx/of/lib/v1alpha1"
	strcase "github.com/cisco-cx/of/wrap/strcase/v1alpha1"
)

// Enforce interface implementation.
func TestInterface(t *testing.T) {
	var c strcase.CaseString = ""
	var _ of.CaseConverter = c

}

// Ensure CaseString is converted from `kebab-case` to `snake_case`.
func TestString_FromKebabToSnake(t *testing.T) {
	// Create an input string without type-inference.
	var c strcase.CaseString = "kebab-case"

	// Capture the return of c.ToSnake() without type-inference.
	var s string = c.ToSnake()

	// Assert that conversion went as planned.
	assert.Equal(t, "kebab_case", s)
}
