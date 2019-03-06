package core

import (
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gr00by87/fst/config"
)

// AllowedRegions stores the list of allowed regions.
var AllowedRegions = []string{"us-east-1", "us-west-2", "eu-west-1", "ap-northeast-1", "ap-southeast-2"}

// Server stores server information data.
type Server struct {
	Name      string
	Env       string
	Type      string
	Region    string
	PrivateIP string
	PublicIP  string
}

// GetAllServers retrieves all servers from given regions and filters them out
// by provided filters.
func GetAllServers(awsCfg config.AWSCredentials, regions []string, filters ...*filter) (servers []Server, err error) {
	for _, region := range regions {
		fromRegion, err := getFromRegion(awsCfg, region, filters)
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
func getFromRegion(awsCfg config.AWSCredentials, region string, filters []*filter) ([]Server, error) {
	creds := credentials.NewStaticCredentials(awsCfg.ID, awsCfg.Secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	svc := ec2.New(session.New(), cfg)

	instances, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, err
	}

	servers := []Server{}
	for _, res := range instances.Reservations {
		server := Server{
			Region:    region,
			PrivateIP: ptrToString(res.Instances[0].PrivateIpAddress),
			PublicIP:  ptrToString(res.Instances[0].PublicIpAddress),
		}

		tags := map[string]*string{
			TagName: &server.Name,
			TagEnv:  &server.Env,
			TagType: &server.Type,
		}

		getTagValues(tags, res.Instances[0].Tags)

		// List only servers with private ip address.
		if server.PrivateIP != "" {
			if checkAllFilters(filters, tags) {
				servers = append(servers, server)
			}
		}
	}

	return servers, nil
}

// getTagValues gets selected tag values from []*ec2.Tag slice.
func getTagValues(tags map[string]*string, ec2Tags []*ec2.Tag) {
	for _, tag := range ec2Tags {
		if val, ok := tags[*tag.Key]; ok {
			*val = *tag.Value
		}
	}
}

// ptrToString returns a string value of a pointer to string.
func ptrToString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}
