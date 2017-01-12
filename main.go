// gotmpl is simple a command line tool that will substitute the variables
// from a data file into a Go text template file.
//
// Get it:
//     go get -u github.com/msample/gotmpl
//
// Use it:
//     gotmpl -d dat.yml cfg.txt.tmpl > cfg.txt
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
	varsFile = flag.String("d", "gotmpl-vars.yml", "YAML, JSON, HCL or TOML file with var values to substitute into the template.")
)

func main() {
	flag.Parse()
	defer glog.Flush()

	var tmpl *template.Template
	var err error

	data, err := parseVars(*varsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading vars data: %v\n", err)
		os.Exit(1)
	}

	glog.Infof("Data is %v\n", data)

	if len(flag.Args()) > 0 {
		tmpl, err = template.ParseFiles(flag.Args()...)
	} else {
		s, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			glog.Errorf("Error reading stdin: %v\n", err)
			os.Exit(2)
		}
		tmpl, err = template.New("t1").Parse(string(s))
	}
	if err != nil {
		glog.Errorf("Template parsing error: %v\n", err)
		os.Exit(3)
	}

	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		glog.Errorf("Template execution error: %v\n", err)
		os.Exit(4)
	}
}

// parseVars tries to parse the input and returnes the first result
// that parses successfully in this order: YAML, JSON, HCL & TOML
func parseVars(file string) (map[string]interface{}, error) {

	// adaptive variant of parse code in github.com/spf13/viper (MIT)
	// minus properties file support

	f, err := os.Open(file)
	if err != nil {
		return nil, err
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

	err = json.Unmarshal(v, &pv)
	if err == nil {
		// yaml should cover JSON but in case...
		glog.Info("JSON detected")
		return pv, nil
	}
	glog.Infof("JSON err: %v", err)

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
