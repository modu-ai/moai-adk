package cli

// RED-phase contract stub for the F1 factory commands (t1239 M5): it fixes the
// clock seam the command tests set. The GREEN commit deletes this file and
// declares the seam next to the commands.

import "time"

var factoryCardNow = time.Now
