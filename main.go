package main

import (
	"github.com/iancmcc/packer-post-processor-ovftool/ovftool"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServePostProcessor(new(ovftool.OVFPostProcessor))
}
