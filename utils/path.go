package utils

import (
	"log"
	"runtime"
)

func GetHostPath() string {
	switch runtime.GOOS {
	case "darwin":
		return ""
	case "linux":
		return ""
	default:
		log.Fatalln("Unsupported OS: ", runtime.GOOS)
		return ""
	}
}
