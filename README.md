Packer ovftool post-processor
=============================

Exposes VMware's [OVF Tool](https://www.vmware.com/support/developer/ovf/) as
a [Packer](http://www.packer.io) post-processor, enabling VMware builds to
produce OVA/OVF artifacts.

For the sake of simplicity, only a few of ovftool's options are currently
exposed, but adding more would be a simple task. Contributions are welcome.

Usage
-----
Add the post-processor to your packer template:

    {
        "post-processors": [{
            "type": "ovftool",
            "only": ["vmware"],
            "format": "ova"
        }]
    }

Available configuration options:

* `target`: The path where the artifact should be created, without the file 
  extension. This is a [configuration template](http://www.packer.io/docs/templates/configuration-templates.html). 
  Defaults to `packer_{{.BuildName}}_{{.Provider}}`.
* `format`:      The target type to create, either "ova" or "ovf". Defaults 
  to "ovf" if not specified.
* `compression`: The compression level to use when creating the artifact. A 
  number from 1-9. Default value is 0, meaning no compression.


Installation
------------
Run:

    $ go get github.com/iancmcc/packer-post-processor-ovftool
    $ go install github.com/iancmcc/packer-post-processor-ovftool

Add the post-processor to ~/.packerconfig:

    {
      "post-processors": {
        "ovftool": "packer-post-processor-ovftool"
      }
    }

### Packer API differences
If you want to use the plugin with Packer v0.4, you'll need to build for API
version 1. Perform installation as above (you can skip the "go install" step), then:

    $ cd $GOPATH/src/github.com/iancmcc/packer-post-processor-ovftool && git checkout v0.4.1
    $ cd $GOPATH/src/github.com/mitchellh/packer && git checkout v0.4.1
    $ go install github.com/iancmcc/packer-post-processor-ovftool

