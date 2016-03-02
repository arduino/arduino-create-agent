//
// Copyright 2014 Cristian Maglie. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//

package serial

// termios manipulation functions

func termiosMask(data int) uint64 {
	return uint64(data)
}
