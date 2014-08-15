package ovftool

import (
	"bytes"
	"errors"
	"fmt"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"strings"
)

var executable string = "ovftool"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	TargetPath          string `mapstructure:"target"`
	TargetType          string `mapstructure:"format"`
	Compression         uint   `mapstructure:"compression"`
	tpl                 *packer.ConfigTemplate
}

type OVFPostProcessor struct {
	cfg Config
}

type OutputPathTemplate struct {
	ArtifactId string
	BuildName  string
	Provider   string
}

func (p *OVFPostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.cfg, raws...)
	if err != nil {
		return err
	}
	p.cfg.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.cfg.tpl.UserVars = p.cfg.PackerUserVars

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
			errs, fmt.Errorf("Error: Could not find ovftool executable: %s", err))
	}

	if err = p.cfg.tpl.Validate(p.cfg.TargetPath); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing target template: %s", err))
	}

	if !(p.cfg.TargetType == "ovf" || p.cfg.TargetType == "ova") {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Invalid target type. Only 'ovf' or 'ova' are allowed."))
	}

	if !(p.cfg.Compression >= 0 && p.cfg.Compression <= 9) {
		errs = packer.MultiErrorAppend(
			errs, errors.New("Invalid compression level. Must be between 1 and 9, or 0 for no compression."))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *OVFPostProcessor) stripDrives(vmx string) error {
	vmxData, err := vmwcommon.ReadVMX(vmx)
	if err != nil {
		return err
	}
	for k, _ := range vmxData {
		if strings.HasPrefix(k, "floppy0.") {
			delete(vmxData, k)
		}
		if strings.HasPrefix(k, "ide1:0.file") {
			delete(vmxData, k)
		}
	}
	vmxData["floppy0.present"] = "FALSE"
	vmxData["ide1:0.present"] = "FALSE"
	if err := vmwcommon.WriteVMX(vmx, vmxData); err != nil {
		return err
	}
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

	// Strip DVD and floppy drives from the VMX
	if err := p.stripDrives(vmx); err != nil {
		return nil, false, fmt.Errorf("Couldn't strip floppy/DVD drives from VMX")
	}

	targetPath, err := p.cfg.tpl.Process(p.cfg.TargetPath, &OutputPathTemplate{
		ArtifactId: artifact.Id(),
		BuildName:  p.cfg.PackerBuildName,
		Provider:   "vmware",
	})
	if err != nil {
		return nil, false, err
	}

	// build the arguments
	args := []string{
		"--targetType=" + p.cfg.TargetType,
		"--acceptAllEulas",
	}

	// append --compression, if it is set
	if p.cfg.Compression > 0 {
		args = append(args, fmt.Sprintf("--compress=%d", p.cfg.Compression))
	}

	// add the source/target
	args = append(args, vmx, targetPath)

	ui.Message(fmt.Sprintf("Executing ovftool with arguments: %+v", args))
	cmd := exec.Command(executable, args...)
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	cmd.Stderr = &buffer
	err = cmd.Run()
	if err != nil {
		return nil, false, fmt.Errorf("Unable to execute ovftool: %s", buffer.String())
	}
	ui.Message(fmt.Sprintf("%s", buffer.String()))

	return artifact, false, nil
}
