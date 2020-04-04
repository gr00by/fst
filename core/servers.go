package core

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/gr00by87/fst/config"
)

const (
	idName      = "tag:Name"
	idPrivateIP = "private-ip-address"
	idPublicIP  = "ip-address"
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

// serverID stores the server identifier and it's type.
type serverID struct {
	typ string
	id  string
}

// NewServerID creates a new serverID.
func NewServerID(id string) serverID {
	sid := serverID{
		id:  id,
		typ: idName,
	}

	ip := net.ParseIP(id)
	if ip != nil {
		if strings.HasPrefix(id, "172.") {
			sid.typ = idPrivateIP
		} else {
			sid.typ = idPublicIP
		}
	}

	return sid
}

// GetAllServers retrieves all servers from given regions and filters them out
// by provided filters.
func GetAllServers(awsCfg config.AWSCredentials, regions []string, filters ...*filter) (servers []Server, err error) {
	for _, region := range regions {
		fromRegion, err := getFromRegion(awsCfg, region, &ec2.DescribeInstancesInput{}, filters...)
		if err != nil {
			return nil, err
		}

		servers = append(servers, fromRegion...)
	}

	sort.Slice(servers, func(i, j int) bool {
		if servers[i].Env == servers[j].Env {
			return strings.ToLower(servers[i].Name) < strings.ToLower(servers[j].Name)
		}
		return servers[i].Env < servers[j].Env
	})

	return
}

// GetSingleServer tries to find a server iterating over all available regions.
// Returns an error if no server is found.
func GetSingleServer(awsCfg config.AWSCredentials, sid serverID) (*Server, error) {
	for _, region := range AllowedRegions {

		servers, err := getFromRegion(awsCfg, region, &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				&ec2.Filter{
					Name: aws.String(sid.typ),
					Values: []*string{
						aws.String(sid.id),
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		if len(servers) > 0 {
			return &servers[0], nil
		}
	}

	return nil, fmt.Errorf("server not found: %s", sid.id)
}

// getFromRegion retrieves servers from a given region and filters them out
// by provided filters.
func getFromRegion(awsCfg config.AWSCredentials, region string, dii *ec2.DescribeInstancesInput, filters ...*filter) ([]Server, error) {
	creds := credentials.NewStaticCredentials(awsCfg.ID, awsCfg.Secret, "")
	cfg := aws.NewConfig().WithRegion(region).WithCredentials(creds)
	svc := ec2.New(session.New(), cfg)

	instances, err := svc.DescribeInstances(dii)
	if err != nil {
		return nil, err
	}

	servers := []Server{}
	for _, res := range instances.Reservations {
		for _, instance := range res.Instances {
			server := Server{
				Region:    region,
				PrivateIP: ptrToString(instance.PrivateIpAddress),
				PublicIP:  ptrToString(instance.PublicIpAddress),
			}

			tags := map[string]*string{
				TagName: &server.Name,
				TagEnv:  &server.Env,
				TagType: &server.Type,
			}

			getTagValues(tags, instance.Tags)

			// List only servers with private ip address.
			if server.PrivateIP != "" {
				if checkAllFilters(filters, tags) {
					servers = append(servers, server)
				}
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
