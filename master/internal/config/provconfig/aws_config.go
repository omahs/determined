package provconfig

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"

	"github.com/determined-ai/determined/master/pkg"
	"github.com/determined-ai/determined/master/pkg/check"
	"github.com/determined-ai/determined/master/pkg/device"
)

//go:generate go run aws_gen.go provconfig ec2InstanceSlots map[Ec2InstanceType]int aws_slots.go
//go:generate gofmt -w aws_slots.go

// SpotPriceNotSetPlaceholder set placeholder.
const SpotPriceNotSetPlaceholder = "OnDemand"

// AWSClusterConfig describes the configuration for an EC2 cluster managed by Determined.
type AWSClusterConfig struct {
	Region string `json:"region"`

	RootVolumeSize int    `json:"root_volume_size"`
	ImageID        string `json:"image_id"`

	TagKey       string `json:"tag_key"`
	TagValue     string `json:"tag_value"`
	InstanceName string `json:"instance_name"`

	SSHKeyName            string              `json:"ssh_key_name"`
	NetworkInterface      ec2NetworkInterface `json:"network_interface"`
	IamInstanceProfileArn string              `json:"iam_instance_profile_arn"`

	InstanceType  Ec2InstanceType `json:"instance_type"`
	InstanceSlots *int            `json:"instance_slots,omitempty"`

	LogGroup  string `json:"log_group"`
	LogStream string `json:"log_stream"`

	SpotEnabled  bool   `json:"spot"`
	SpotMaxPrice string `json:"spot_max_price"`

	CustomTags []*ec2Tag `json:"custom_tags"`

	CPUSlotsAllowed bool `json:"cpu_slots_allowed"`
}

var defaultAWSImageID = map[string]string{
	"ap-northeast-1": "ami-0efbc837b3c729df1",
	"ap-northeast-2": "ami-0934d35fc17d76abc",
	"ap-southeast-1": "ami-05068dcfa829e229e",
	"ap-southeast-2": "ami-0e62ea531c3969a65",
	"us-east-2":      "ami-0ba898a163ad80bf9",
	"us-east-1":      "ami-0b4cc2ed4ecd07a2d",
	"us-west-2":      "ami-0a483013f2e53ed60",
	"eu-central-1":   "ami-051a91b2829200200",
	"eu-west-2":      "ami-08aa94155b641d19d",
	"eu-west-1":      "ami-04eab4dc55258e621",
}

var defaultAWSClusterConfig = AWSClusterConfig{
	InstanceName:   "determined-ai-agent",
	RootVolumeSize: 200,
	TagKey:         "managed_by",
	NetworkInterface: ec2NetworkInterface{
		PublicIP: true,
	},
	InstanceType:    "g4dn.metal",
	SpotEnabled:     false,
	CPUSlotsAllowed: false,
}

// BuildDockerLogString build docker log string.
func (c *AWSClusterConfig) BuildDockerLogString() string {
	logString := ""
	if c.LogGroup != "" {
		logString += "--log-driver=awslogs --log-opt awslogs-group=" + c.LogGroup
	}
	if c.LogStream != "" {
		logString += " --log-opt awslogs-stream=" + c.LogStream
	}
	return logString
}

// InitDefaultValues init default values.
func (c *AWSClusterConfig) InitDefaultValues() error {
	metadata, err := getEC2MetadataSess()
	if err != nil {
		return err
	}

	if len(c.Region) == 0 {
		if c.Region, err = metadata.Region(); err != nil {
			return err
		}
	}

	if len(c.SpotMaxPrice) == 0 {
		c.SpotMaxPrice = SpotPriceNotSetPlaceholder
	}

	if len(c.ImageID) == 0 {
		if v, ok := defaultAWSImageID[c.Region]; ok {
			c.ImageID = v
		} else {
			return errors.Errorf("cannot find default image ID in the region %s", c.Region)
		}
	}

	// One common reason that metadata.GetInstanceIdentityDocument() fails is that the master is not
	// running in EC2. Use a default name here rather than holding up initializing the provider.
	identifier := pkg.DeterminedIdentifier
	idDoc, err := metadata.GetInstanceIdentityDocument()
	if err == nil {
		identifier = idDoc.InstanceID
	}

	if len(c.TagValue) == 0 {
		c.TagValue = identifier
	}
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *AWSClusterConfig) UnmarshalJSON(data []byte) error {
	*c = defaultAWSClusterConfig
	type DefaultParser *AWSClusterConfig
	return json.Unmarshal(data, DefaultParser(c))
}

func validateInstanceTypeSlots(c AWSClusterConfig) error {
	// Must have an instance in ec2InstanceSlots map or InstanceSlots set
	instanceType := c.InstanceType
	if _, ok := ec2InstanceSlots[instanceType]; ok {
		return nil
	}

	instanceSlots := c.InstanceSlots
	if instanceSlots != nil {
		if *instanceSlots < 0 {
			return errors.Errorf("ec2 'instance_slots' must be greater than or equal to 0")
		}
		ec2InstanceSlots[instanceType] = *instanceSlots
		return nil
	}

	strs := make([]string, 0, len(ec2InstanceSlots))
	for t := range ec2InstanceSlots {
		strs = append(strs, t.Name())
	}
	return errors.Errorf("Either ec2 'instance_type' and 'instance_slots' must be specified or "+
		"the ec2 'instance_type' must be one of types: %s", strings.Join(strs, ", "))
}

// Validate implements the check.Validatable interface.
func (c AWSClusterConfig) Validate() []error {
	var spotPriceIsNotValidNumberErr error
	if c.SpotEnabled && c.SpotMaxPrice != SpotPriceNotSetPlaceholder {
		spotPriceIsNotValidNumberErr = validateMaxSpotPrice(c.SpotMaxPrice)
	}
	return []error{
		check.GreaterThan(len(c.SSHKeyName), 0, "ec2 key name must be non-empty"),
		check.GreaterThanOrEqualTo(c.RootVolumeSize, 100, "ec2 root volume size must be >= 100"),
		spotPriceIsNotValidNumberErr,
		validateInstanceTypeSlots(c),
	}
}

// SlotsPerInstance returns the number of slots per instance.
func (c AWSClusterConfig) SlotsPerInstance() int {
	slots := c.InstanceType.Slots()
	if slots == 0 && c.CPUSlotsAllowed {
		slots = 1
	}

	return slots
}

// SlotType returns the type of the slot.
func (c AWSClusterConfig) SlotType() device.Type {
	slots := c.InstanceType.Slots()
	if slots > 0 {
		return device.CUDA
	}
	if c.CPUSlotsAllowed {
		return device.CPU
	}
	return device.ZeroSlot
}

// Accelerator returns the GPU accelerator for the instance.
func (c AWSClusterConfig) Accelerator() string {
	return c.InstanceType.Accelerator()
}

func validateMaxSpotPrice(spotMaxPriceInput string) error {
	// Must have 1 or 0 decimalPoints. All other characters must be digits
	numDecimalPoints := strings.Count(spotMaxPriceInput, ".")
	if numDecimalPoints != 0 && numDecimalPoints != 1 {
		return errors.New(
			fmt.Sprintf("spot max price should have either 0 or 1 decimal points. "+
				"Received %s, which has %d decimal points",
				spotMaxPriceInput,
				numDecimalPoints))
	}

	priceWithoutDecimalPoint := strings.ReplaceAll(spotMaxPriceInput, ".", "")
	for _, char := range priceWithoutDecimalPoint {
		if !unicode.IsDigit(char) {
			return errors.New(
				fmt.Sprintf("spot max price should only contain digits and, optionally, one decimal point. "+
					"Received %s, which has the non-digit character %s",
					spotMaxPriceInput,
					string(char)))
		}
	}
	return nil
}

type ec2NetworkInterface struct {
	PublicIP        bool   `json:"public_ip"`
	SubnetID        string `json:"subnet_id"`
	SecurityGroupID string `json:"security_group_id"`
}

type ec2Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Ec2InstanceType is Ec2InstanceType.
type Ec2InstanceType string

// Name returns the string representation of instance type.
func (t Ec2InstanceType) Name() string {
	return string(t)
}

// Slots returns number of slots.
func (t Ec2InstanceType) Slots() int {
	if s, ok := ec2InstanceSlots[t]; ok {
		return s
	}
	return 0
}

// Accelerator source:
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/accelerated-computing-instances.html
func (t Ec2InstanceType) Accelerator() string {
	instanceType := t.Name()
	numGpu := t.Slots()
	accelerator := ""
	if strings.HasPrefix(instanceType, "p2") {
		accelerator = "NVIDIA Tesla K80"
	}
	if strings.HasPrefix(instanceType, "p3") {
		accelerator = "NVIDIA Tesla V100"
	}
	if strings.HasPrefix(instanceType, "p4d") {
		accelerator = "NVIDIA A100"
	}
	if strings.HasPrefix(instanceType, "g3") {
		accelerator = "NVIDIA Tesla M60"
	}
	if strings.HasPrefix(instanceType, "g5g") {
		accelerator = "NVIDIA T4G"
	}
	if strings.HasPrefix(instanceType, "g5") {
		accelerator = "NVIDIA A10G"
	}
	if strings.HasPrefix(instanceType, "g4dn") {
		accelerator = "NVIDIA T4 Tensor Core"
	}
	if accelerator == "" {
		return ""
	}
	return fmt.Sprintf("%d x %s", numGpu, accelerator)
}

func getEC2MetadataSess() (*ec2metadata.EC2Metadata, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AWS session")
	}
	return ec2metadata.New(sess), nil
}

func getEC2Metadata(field string) (string, error) {
	ec2Metadata, err := getEC2MetadataSess()
	if err != nil {
		return "", err
	}
	return ec2Metadata.GetMetadata(field)
}

func onEC2() bool {
	ec2Metadata, err := getEC2MetadataSess()
	if err != nil {
		return false
	}
	return ec2Metadata.Available()
}
