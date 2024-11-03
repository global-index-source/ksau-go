// SPDX-License-Identifier: Apache-2.0

package internal

import "runtime"

func GetFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}
