// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package watcher

import "os"

// interruptSignal returns the signal used to request a clean shutdown. On Linux this is SIGINT.
func interruptSignal() os.Signal { return os.Interrupt }
