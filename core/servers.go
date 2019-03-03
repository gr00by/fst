package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gr00by87/fst/config"
)

// allowedRegions stores the list of allowed regions.
var allowedRegions = []string{"us-east-1", "us-west-2", "eu-west-1", "ap-northeast-1", "ap-southeast-2"}

// Server stores server information data.
type Server struct {
	Name    string
	Env     string
	Address string
}

// GetAllServers retrieves all servers from given regions and filters them out
// by provided filters.
func GetAllServers(awsCfg config.AWSCredentials, names, envs *filter, regions []string) (servers []Server, err error) {
	regions, err = checkRegions(regions)
	if err != nil {
		return
	}

	for _, region := range regions {
		fromRegion, err := getFromRegion(awsCfg, names, envs, region)
		if err != nil {
			return nil, err
		}

		servers = append(servers, fromRegion...)
	}

	sort.Slice(servers, func(i, j int) bool {
		if servers[i].Env == servers[j].Env {
			return strings.ToLower(servers[i].Name) < strings.ToLower(servers[j].Name)
		} else {
			return servers[i].Env < servers[j].Env
		}
	})

	return
}

// getFromRegion retrieves servers from a given region and filters them out
// by provided filters.
func getFromRegion(awsCfg config.AWSCredentials, names, envs *filter, region string) ([]Server, error) {
	creds := credentials.NewStaticCredentials(awsCfg.ID, awsCfg.Secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	svc := ec2.New(session.New(), cfg)

	instances, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	servers := []Server{}
	for _, res := range instances.Reservations {
		name := getTagValue("Name", res.Instances[0].Tags)
		env := getTagValue("Env", res.Instances[0].Tags)

		if ip := res.Instances[0].PublicIpAddress; ip != nil && name != "" && env != "" {
			if names.compareValue(name) && envs.compareValue(env) {
				servers = append(servers, Server{
					Name:    name,
					Env:     env,
					Address: *ip,
				})
			}
		}
	}

	return servers, nil
}

// checkRegions checks if given regions are allowed.
func checkRegions(regions []string) ([]string, error) {
	if regions[0] == "all" {
		return allowedRegions, nil
	}

	allowedRegionsFilter := NewFilter(allowedRegions, Equals, DontIgnoreCase)
	for _, region := range regions {
		if !allowedRegionsFilter.compareValue(region) {
			return nil, fmt.Errorf("invalid region: %s", region)
		}
	}

	return regions, nil
}

// getTagValues gets tag value from []*ec2.Tag slice.
func getTagValue(key string, tags []*ec2.Tag) string {
	for _, tag := range tags {
		if *tag.Key == key {
			return *tag.Value
		}
	}

	return ""
}
