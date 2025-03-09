package main

import (
	"flag"
	"fmt"
	"runtime/debug"
)

// version is intended to be used with the -ldflags switch.
var version = ""

const helpText = `sms is a minimalistic mock http server that uses a filesystem as backend

Usage:

  --help	print help
  --version	print version

Environment Variables:

  PORT (default: 4321)
  ADDRESS (default: ":$PORT")
  LOG_LEVEL (default: "debug"")
  RESPONSES_DIR - Directory where the response files are located (default: "./.sms_responses")
  EXTENSION_MIME_TYPE_MAP - File extension to http request Accept MIME type, e.g. "txt:text/plain"
  METHOD_STATUS_MAP - Request http method to response http status (default: "DELETE:202,GET:200,PATCH:204,POST:201,PUT:204")

`

func processFlags() (stop bool) {
	v := flag.Bool("version", false, "print version")
	h := flag.Bool("help", false, "print help")

	flag.Parse()

	if v != nil && *v {
		if version != "" {
			fmt.Println(version)
		}

		if bi, ok := debug.ReadBuildInfo(); ok {
			fmt.Println(bi.Main.Version)
		}

		return true
	}

	if h != nil && *h {
		fmt.Print(helpText)
		return true
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("couldn't read build info")
	}

	fmt.Printf("%s version %s\n", bi.Path, bi.Main.Version)

	return false
}
