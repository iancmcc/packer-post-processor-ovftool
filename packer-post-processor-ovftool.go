package main

import (
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"os/exec"
	"strings"
)

var executable string = "ovftool"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	TargetPath          string `mapstructure:"target"`
	TargetType          string `mapstructure:"format"`
	Compression         uint   `mapstructure:"compression"`
}

type OVFPostProcessor struct {
	cfg Config
}

func (p *OVFPostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.cfg, raws...)
	if err != nil {
		return err
	}
	tpl, err := packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	tpl.UserVars = p.cfg.PackerUserVars

	if p.cfg.TargetType == "" {
		p.cfg.TargetType = "ovf"
	}

	if p.cfg.TargetPath == "" {
		p.cfg.TargetPath = "packer_{{ .BuildName }}_{{.Provider}}"
		if p.cfg.TargetType == "ova" {
			p.cfg.TargetPath += ".ova"
		}
	}

	errs := new(packer.MultiError)

	_, err = exec.LookPath(executable)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error: Could not find ovftool executable.", err))
	}

	p.cfg.TargetPath, err = tpl.Process(p.cfg.TargetPath, nil)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing target template: %s", err))
	}

	if !(p.cfg.TargetType == "ovf" || p.cfg.TargetType == "ova") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Invalid target type. Only 'ovf' or 'ova' are allowed."))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *OVFPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != "mitchellh.vmware" {
		return nil, false, fmt.Errorf("ovftool post-processor can only be used on VMware boxes: %s", artifact.BuilderId())
	}

	vmx := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".vmx") {
			vmx = path
		}
	}
	if vmx == "" {
		return nil, false, fmt.Errorf("VMX file could not be located.")
	}

	cmd := exec.Command(
		executable,
		"--targetType="+p.cfg.TargetType,
		"--acceptAllEulas",
		vmx,
		p.cfg.TargetPath)

	err := cmd.Run()
	if err != nil {
		return nil, false, fmt.Errorf("Unable to execute ovftool.", nil)
	}

	return artifact, false, nil
}

func main() {
	plugin.ServePostProcessor(new(OVFPostProcessor))
}
