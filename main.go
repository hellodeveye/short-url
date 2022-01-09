package main

func main() {
	app := App{}
	app.Initialize(getEnvConfig())
	app.Run(":8000")
}
