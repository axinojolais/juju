// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"fmt"

	"github.com/juju/cmd"
	"github.com/juju/errors"
	"github.com/juju/utils/series"
	"launchpad.net/gnuflag"

	"github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/cmd/envcmd"
)

func newAddImageMetadataCommand() cmd.Command {
	return envcmd.Wrap(&addImageMetadataCommand{})
}

const addImageCommandDoc = `
Add image metadata to Juju environment.

Image metadata properties vary between providers. Consequently, some properties
are optional for this command but they may still be needed by your provider.

This command takes only one positional argument - an image id.

arguments:
image-id
   image identifier

options:
-e, --environment (= "")
   juju environment to operate in
--region
   cloud region
--series (= "trusty")
   image series
--arch (= "amd64")
   image architectures
--virt-type
   virtualisation type [provider specific], e.g. hmv
--storage-type
   root storage type [provider specific], e.g. ebs
--storage-size
   root storage size [provider specific]
--stream (= "released")
   image stream

`

// addImageMetadataCommand stores image metadata in Juju environment.
type addImageMetadataCommand struct {
	cloudImageMetadataCommandBase

	ImageId         string
	Region          string
	Series          string
	Arch            string
	VirtType        string
	RootStorageType string
	RootStorageSize uint64
	Stream          string
	Version         string
}

// Init implements Command.Init.
func (c *addImageMetadataCommand) Init(args []string) (err error) {
	if len(args) == 0 {
		return errors.New("image id must be supplied when adding image metadata")
	}
	if len(args) != 1 {
		return errors.New("only one image id can be supplied as an argument to this command")
	}
	c.ImageId = args[0]
	return c.validate()
}

// Info implements Command.Info.
func (c *addImageMetadataCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "add-image",
		Purpose: "adds image metadata to environment",
		Doc:     addImageCommandDoc,
	}
}

// SetFlags implements Command.SetFlags.
func (c *addImageMetadataCommand) SetFlags(f *gnuflag.FlagSet) {
	c.cloudImageMetadataCommandBase.SetFlags(f)

	f.StringVar(&c.Region, "region", "", "image cloud region")
	// TODO (anastasiamac 2015-09-30) Ideally default should be latest LTS.
	// Hard-coding "trusty" for now.
	f.StringVar(&c.Series, "series", "trusty", "image series")
	f.StringVar(&c.Arch, "arch", "amd64", "image architecture")
	f.StringVar(&c.VirtType, "virt-type", "", "image metadata virtualisation type")
	f.StringVar(&c.RootStorageType, "storage-type", "", "image metadata root storage type")
	f.Uint64Var(&c.RootStorageSize, "storage-size", 0, "image metadata root storage size")
	f.StringVar(&c.Stream, "stream", "released", "image metadata stream")
}

// Run implements Command.Run.
func (c *addImageMetadataCommand) Run(ctx *cmd.Context) (err error) {
	api, err := getImageMetadataAddAPI(c)
	if err != nil {
		return err
	}
	defer api.Close()

	m := c.constructMetadataParam()
	found, err := api.Save([]params.CloudImageMetadata{m})
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return nil
	}
	if len(found) > 1 {
		return errors.New(fmt.Sprintf("expected one result, got %d", len(found)))
	}
	if found[0].Error != nil {
		return errors.New(found[0].Error.GoString())
	}
	return nil
}

// MetadataAddAPI defines the API methods that add image metadata command uses.
type MetadataAddAPI interface {
	Close() error
	Save(metadata []params.CloudImageMetadata) ([]params.ErrorResult, error)
}

var getImageMetadataAddAPI = (*addImageMetadataCommand).getImageMetadataAddAPI

func (c *addImageMetadataCommand) getImageMetadataAddAPI() (MetadataAddAPI, error) {
	return c.NewImageMetadataAPI()
}

// Init implements Command.Init.
func (c *addImageMetadataCommand) validate() error {
	v, err := series.SeriesVersion(c.Series)
	if err != nil {
		return errors.Trace(err)
	}
	c.Version = v
	return nil
}

// constructMetadataParam returns cloud image metadata as a param.
func (c *addImageMetadataCommand) constructMetadataParam() params.CloudImageMetadata {
	info := params.CloudImageMetadata{
		ImageId:         c.ImageId,
		Region:          c.Region,
		Version:         c.Version,
		Arch:            c.Arch,
		VirtType:        c.VirtType,
		RootStorageType: c.RootStorageType,
		Stream:          c.Stream,
		Source:          "custom",
	}
	if c.RootStorageSize != 0 {
		info.RootStorageSize = &c.RootStorageSize
	}
	return info
}
