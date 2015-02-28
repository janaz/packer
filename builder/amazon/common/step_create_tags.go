package common

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCreateTags struct {
	Tags map[string]string
}

func (s *StepCreateTags) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)

	if len(s.Tags) > 0 {
		for region, ami := range amis {
			ui.Say(fmt.Sprintf("Adding tags to AMI (%s)...", ami))

			var ec2Tags []ec2.Tag
			for key, value := range s.Tags {
				ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"", key, value))
				ec2Tags = append(ec2Tags, ec2.Tag{key, value})
			}

			regionconn := ec2.New(ec2conn.Auth, aws.Regions[region])

			imagesResp, err := ec2conn.Images([]string{ami}, nil)
			if err != nil {
				err := fmt.Errorf("Error searching for AMI (%s) in %s: %s", ami, region, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			image := &imagesResp.Images[0]

			resourcesToTag := []string{ami}
			for _, blockDevice := range image.BlockDevices {
				resourcesToTag = append(resourcesToTag, blockDevice.SnapshotId)
			}

			_, err = regionconn.CreateTags(resourcesToTag, ec2Tags)
			if err != nil {
				err := fmt.Errorf("Error adding tags to AMI (%s): %s", ami, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateTags) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
