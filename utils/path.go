package utils

import (
	"log"
	"runtime"
)

func GetHostPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	case "linux":
		return "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	default:
		log.Fatalln("Unsupported OS: ", runtime.GOOS)
		return ""
	}
}
