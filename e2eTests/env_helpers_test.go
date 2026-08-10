//     MIT License
//
//     Copyright (c) Microsoft Corporation.
//
//     Permission is hereby granted, free of charge, to any person obtaining a copy
//     of this software and associated documentation files (the "Software"), to deal
//     in the Software without restriction, including without limitation the rights
//     to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//     copies of the Software, and to permit persons to whom the Software is
//     furnished to do so, subject to the following conditions:
//
//     The above copyright notice and this permission notice shall be included in all
//     copies or substantial portions of the Software.
//
//     THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//     IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//     FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//     AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//     LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//     OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//     SOFTWARE

package e2etests

import "testing"

func TestMpfEnv_LegacyFirstThenSnakeCase(t *testing.T) {
	const (
		legacy = "MPF_TESTHELPER_LEGACY"
		snake  = "MPF_TESTHELPER_SNAKE"
	)
	t.Setenv(legacy, "")
	t.Setenv(snake, "")
	t.Setenv(legacy, "legacy-value")
	t.Setenv(snake, "snake-value")

	if got := mpfEnv(legacy, snake); got != "legacy-value" {
		t.Errorf("mpfEnv() = %q, want legacy value %q", got, "legacy-value")
	}
}

func TestMpfEnv_FallsBackToSnakeCase(t *testing.T) {
	const (
		legacy = "MPF_TESTHELPER_LEGACY_EMPTY"
		snake  = "MPF_TESTHELPER_SNAKE_ONLY"
	)
	t.Setenv(legacy, "")
	t.Setenv(snake, "snake-only")

	if got := mpfEnv(legacy, snake); got != "snake-only" {
		t.Errorf("mpfEnv() = %q, want snake_case value %q", got, "snake-only")
	}
}

func TestMpfEnv_EmptyWhenUnset(t *testing.T) {
	const missing = "MPF_TESTHELPER_MISSING"
	t.Setenv(missing, "")

	if got := mpfEnv(missing); got != "" {
		t.Errorf("mpfEnv() = %q, want empty string", got)
	}
}
