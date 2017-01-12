// gotmpl is a simple command line tool that will substitute the
// variables from a data file into a Go text template file.
//
// Get it:
//     go get -u github.com/msample/gotmpl
//
// Use it:
//     gotmpl -d dat.yml cfg.txt.tmpl > cfg.txt
//
//     cat dat.yml | gotmpl cfg.txt.tmpl > cfg.txt
//
//     cat cfg.txt.tmpl | gotmpl -d data.json > cfg.txt
//
//     gotmpl -d dat.yml cfg.txt.tmpl cfg.txt.tmpl2 > cfg.txt
//
//     gotmpl -logtostderr -d dat.yml cfg.txt.tmpl > cfg.txt
//
//     gotmpl -h
//
//
// Data file may contain YAML, JSON, HCL or TOML. Gotmpl tries the
// parsers in that order and takes the result of the first one that
// doesn't complain.
//
// Use -logtostderr option if having problems. Template syntax defined
// here: https://godoc.org/text/template
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/golang/glog"
	"github.com/hashicorp/hcl"
	toml "github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
)

var (
	varsFile = flag.String("d", "-", "YAML, JSON, HCL or TOML file with var values to substitute into the template. Use '-' for stdin (default).")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [tmplateFile...]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()
	defer glog.Flush()

	if len(flag.Args()) == 0 && *varsFile == "-" {
		fmt.Fprintf(os.Stderr, "Cannot read both template and data from stdin\n")
		Usage()
		os.Exit(1)
	}

	var tmpl *template.Template
	var data map[string]interface{}
	var err error

	// read stdin last so fail fast&first file-based info
	if *varsFile == "-" {
		tmpl = readTemplates()
		data = readData()
	} else {
		data = readData()
		tmpl = readTemplates()
	}

	glog.Infof("Data is %v\n", data)

	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		glog.Errorf("Template execution error: %v\n", err)
		os.Exit(2)
	}
}

func readData() map[string]interface{} {
	data, err := parseVars(*varsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading vars data: %v\n", err)
		os.Exit(3)
	}
	return data
}

func readTemplates() *template.Template {

	// files
	if len(flag.Args()) > 0 {
		tmpl, err := template.ParseFiles(flag.Args()...)
		if err != nil {
			glog.Errorf("Template parsing error: %v\n", err)
			os.Exit(4)
		}
		return tmpl
	}

	// stdin
	s, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		glog.Errorf("Error reading stdin: %v\n", err)
		os.Exit(5)
	}
	tmpl, err := template.New("t1").Parse(string(s))

	if err != nil {
		glog.Errorf("Template parsing error: %v\n", err)
		os.Exit(6)
	}
	return tmpl
}

// parseVars tries to parse the input and returns the result of the
// first successful parse in this order: YAML, JSON, HCL & TOML.
// File value of "-" will read stdin until closed and then parse.
func parseVars(file string) (map[string]interface{}, error) {

	// adaptive variant of parse code in github.com/spf13/viper (MIT)
	// minus properties file support
	f := os.Stdin
	var err error
	if file != "-" {
		f, err = os.Open(file)
		if err != nil {
			return nil, err
		}
	}

	v, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	pv := make(map[string]interface{})
	err = yaml.Unmarshal(v, &pv)
	if err == nil {
		glog.Info("Yaml detected")
		return pv, nil
	}
	glog.Infof("YAML err: %v", err)

	pv = make(map[string]interface{})
	err = json.Unmarshal(v, &pv)
	if err == nil {
		// yaml should cover JSON but in case...
		glog.Info("JSON detected")
		return pv, nil
	}
	glog.Infof("JSON err: %v", err)

	pv = make(map[string]interface{})
	o, err1 := hcl.Parse(string(v))
	var err2 error
	if err1 == nil {
		err2 = hcl.DecodeObject(&pv, o)
	}
	if err1 == nil && err2 == nil {
		glog.Info("HCL detected")
		return pv, nil
	}
	glog.Infof("HCL errs: %v, %v", err1, err2)

	pv = make(map[string]interface{})
	t, err := toml.LoadReader(bytes.NewBuffer(v))
	if err == nil {
		tm := t.ToMap()
		for k, v := range tm {
			pv[k] = v
		}
		glog.Info("TOML detected")
		return pv, nil
	}
	glog.Infof("TOML err: %v", err)

	return nil, fmt.Errorf("data in '%v' failed to parse as YAML, JSON, HCL or TOML", file)
}
