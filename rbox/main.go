package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/twitchyliquid64/raspberry-box/interpreter"
)

var (
	script   = flag.String("script", "build.box", "Path to the box build file.")
	template = flag.String("img", "", "Path to the base image file.")
	verbose  = flag.Bool("verbose", false, "Enables verbose logging.")
)

func loadScript() ([]byte, error) {
	d, err := os.Stat(*script)
	if err != nil {
		return nil, err
	}
	if d.IsDir() {
		return nil, fmt.Errorf("%v is a directory", *script)
	}
	return ioutil.ReadFile(*script)
}

func main() {
	flag.Parse()
	sData, err := loadScript()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load script: %v\n", err)
		os.Exit(1)
	}

	script, err := interpreter.NewScript(sData, *script, *verbose, &interpreter.WDLoader{}, flag.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initialization failed: %v\n", err)
		os.Exit(1)
	}

	if err := run(script); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(s *interpreter.Script) error {
	defer s.Close()
	if *template == "" {
		var err error
		*template, err = s.CallFn("fallback_template")
		if err != nil {
			return fmt.Errorf("fallback_template() failed: %v", err)
		}
	}

	if err := s.Setup(*template); err != nil {
		return fmt.Errorf("setup() failed: %v", err)
	}
	if err := s.Build(); err != nil {
		return fmt.Errorf("build() failed: %v", err)
	}
	return nil
}
