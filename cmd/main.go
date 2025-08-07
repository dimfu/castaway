package main

func main() {
	if err := newApp().run(); err != nil {
		panic(err)
	}
}
