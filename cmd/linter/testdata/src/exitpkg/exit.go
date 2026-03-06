package exitpkg

import "os"

func Test() {
	os.Exit(1) // want "os.Exit cannot be used outside main"
}
