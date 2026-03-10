package panicpkg

func Test() {
	panic("error") // want "panic usage is forbidden"
}
