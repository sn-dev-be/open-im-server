package club

import (
	"fmt"
	"os"
)

func SyncServerToBusiness() error {
	for i := 0; i < 1000; i++ {
		fmt.Fprintln(os.Stderr, i)
	}
	return nil
}
