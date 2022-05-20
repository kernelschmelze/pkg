```
package main

import (
	log "github.com/kernelschmelze/pkg/logger"
)

func main() {

	log.Info("hallo")
	log.Debug("hallo")

	log.Warnf("warning: %d failed", 4711)

}


// 0721 15:56:49.278492 INF hallo main.go:9
// 0721 15:56:49.278570 DBG hallo main.go:10
// 0721 15:56:49.278583 WRN warning: 4711 failed main.go:12

```