# gotmpl
--
gotmpl is a simple command line tool that will substitute the variables from a
data file into a Go text template file.

Get it:

    go get -u github.com/msample/gotmpl

Use it:

    gotmpl -d dat.yml cfg.txt.tmpl > cfg.txt

    cat dat.yml | gotmpl cfg.txt.tmpl > cfg.txt

    cat cfg.txt.tmpl | gotmpl -d data.json > cfg.txt

    gotmpl -d dat.yml cfg.txt.tmpl cfg.txt.tmpl2 > cfg.txt

    gotmpl -logtostderr -d dat.yml cfg.txt.tmpl > cfg.txt

    gotmpl -h

Data file may contain YAML, JSON, HCL or TOML. Gotmpl tries the parsers in that
order and takes the result of the first one that doesn't complain.

Use -logtostderr option if having problems. Template syntax defined here:
https://godoc.org/text/template
