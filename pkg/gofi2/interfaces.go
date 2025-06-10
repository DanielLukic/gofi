package gofi2

type AppInterface interface {
	Start() error
	Show()
	Hide()
	Run()
	Exit()
	Cleanup()
}
