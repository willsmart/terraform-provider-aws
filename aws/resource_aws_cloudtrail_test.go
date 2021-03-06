package aws

import (
	"fmt"
	"log"
	"testing"

	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudTrail(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Trail": {
			"basic":                      testAccAWSCloudTrail_basic,
			"cloudwatch":                 testAccAWSCloudTrail_cloudwatch,
			"enableLogging":              testAccAWSCloudTrail_enable_logging,
			"includeGlobalServiceEvents": testAccAWSCloudTrail_include_global_service_events,
			"isMultiRegion":              testAccAWSCloudTrail_is_multi_region,
			"isOrganization":             testAccAWSCloudTrail_is_organization,
			"logValidation":              testAccAWSCloudTrail_logValidation,
			"kmsKey":                     testAccAWSCloudTrail_kmsKey,
			"tags":                       testAccAWSCloudTrail_tags,
			"eventSelector":              testAccAWSCloudTrail_event_selector,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAWSCloudTrail_basic(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "include_global_service_events", "true"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfigModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "s3_key_prefix", "/prefix"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "include_global_service_events", "false"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_cloudwatch(t *testing.T) {
	var trail cloudtrail.Trail
	randInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfigCloudWatch(randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.test", &trail),
					resource.TestCheckResourceAttrSet("aws_cloudtrail.test", "cloud_watch_logs_group_arn"),
					resource.TestCheckResourceAttrSet("aws_cloudtrail.test", "cloud_watch_logs_role_arn"),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfigCloudWatchModified(randInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.test", &trail),
					resource.TestCheckResourceAttrSet("aws_cloudtrail.test", "cloud_watch_logs_group_arn"),
					resource.TestCheckResourceAttrSet("aws_cloudtrail.test", "cloud_watch_logs_role_arn"),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_enable_logging(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					// AWS will create the trail with logging turned off.
					// Test that "enable_logging" default works.
					testAccCheckCloudTrailLoggingEnabled("aws_cloudtrail.foobar", true, &trail),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfigModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					testAccCheckCloudTrailLoggingEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					testAccCheckCloudTrailLoggingEnabled("aws_cloudtrail.foobar", true, &trail),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_is_multi_region(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				Config: testAccAWSCloudTrailConfigMultiRegion(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "is_multi_region_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "is_multi_region_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_is_organization(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfigOrganization(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "is_organization_trail", "true"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "is_organization_trail", "false"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_logValidation(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_logValidation(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "s3_key_prefix", ""),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", true, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig_logValidationModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "s3_key_prefix", ""),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_kmsKey(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()
	keyRegex := regexp.MustCompile(`^arn:aws:([a-zA-Z0-9\-])+:([a-z]{2}-[a-z]+-\d{1})?:(\d{12})?:(.*)$`)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_kmsKey(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "s3_key_prefix", ""),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "include_global_service_events", "true"),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "kms_key_id", keyRegex),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSCloudTrail_tags(t *testing.T) {
	var trail cloudtrail.Trail
	var trailTags []*cloudtrail.Tag
	var trailTagsModified []*cloudtrail.Tag
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_tags(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "tags.%", "2"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTags),
					testAccCheckCloudTrailCheckTags(&trailTags, map[string]string{"Foo": "moo", "Pooh": "hi"}),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig_tagsModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "tags.%", "3"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTagsModified),
					testAccCheckCloudTrailCheckTags(&trailTagsModified, map[string]string{"Foo": "moo", "Moo": "boom", "Pooh": "hi"}),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
			{
				Config: testAccAWSCloudTrailConfig_tagsModifiedAgain(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "tags.%", "0"),
					testAccCheckCloudTrailLoadTags(&trail, &trailTagsModified),
					testAccCheckCloudTrailCheckTags(&trailTagsModified, map[string]string{}),
					testAccCheckCloudTrailLogValidationEnabled("aws_cloudtrail.foobar", false, &trail),
					testAccCheckCloudTrailKmsKeyIdEquals("aws_cloudtrail.foobar", "", &trail),
				),
			},
		},
	})
}

func testAccAWSCloudTrail_include_global_service_events(t *testing.T) {
	var trail cloudtrail.Trail
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_include_global_service_events(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudTrailExists("aws_cloudtrail.foobar", &trail),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "include_global_service_events", "false"),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSCloudTrail_event_selector(t *testing.T) {
	cloudTrailRandInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudTrailDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudTrailConfig_eventSelector(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.values.#", "2"),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.values.0", regexp.MustCompile(`^arn:[^:]+:s3:::.+/foobar$`)),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.values.1", regexp.MustCompile(`^arn:[^:]+:s3:::.+/baz$`)),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.include_management_events", "false"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.read_write_type", "ReadOnly"),
				),
			},
			{
				ResourceName:      "aws_cloudtrail.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudTrailConfig_eventSelectorModified(cloudTrailRandInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.#", "2"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.values.#", "2"),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.values.0", regexp.MustCompile(`^arn:[^:]+:s3:::.+/foobar$`)),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "event_selector.0.data_resource.0.values.1", regexp.MustCompile(`^arn:[^:]+:s3:::.+/baz$`)),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.include_management_events", "true"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.0.read_write_type", "ReadOnly"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.#", "2"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.0.type", "AWS::S3::Object"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.0.values.#", "2"),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.0.values.0", regexp.MustCompile(`^arn:[^:]+:s3:::.+/tf1$`)),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.0.values.1", regexp.MustCompile(`^arn:[^:]+:s3:::.+/tf2$`)),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.1.type", "AWS::Lambda::Function"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.1.values.#", "1"),
					resource.TestMatchResourceAttr("aws_cloudtrail.foobar", "event_selector.1.data_resource.1.values.0", regexp.MustCompile(`^arn:[^:]+:lambda:.+:tf-test-trail-event-select-\d+$`)),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.1.include_management_events", "false"),
					resource.TestCheckResourceAttr("aws_cloudtrail.foobar", "event_selector.1.read_write_type", "All"),
				),
			},
		},
	})
}

func testAccCheckCloudTrailExists(n string, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeTrails(&params)
		if err != nil {
			return err
		}
		if len(resp.TrailList) == 0 {
			return fmt.Errorf("Trail not found")
		}
		*trail = *resp.TrailList[0]

		return nil
	}
}

func testAccCheckCloudTrailLoggingEnabled(n string, desired bool, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		params := cloudtrail.GetTrailStatusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTrailStatus(&params)

		if err != nil {
			return err
		}
		if *resp.IsLogging != desired {
			return fmt.Errorf("Expected logging status %t, given %t", desired, *resp.IsLogging)
		}

		return nil
	}
}

func testAccCheckCloudTrailLogValidationEnabled(n string, desired bool, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if trail.LogFileValidationEnabled == nil {
			return fmt.Errorf("No LogFileValidationEnabled attribute present in trail: %s", trail)
		}

		if *trail.LogFileValidationEnabled != desired {
			return fmt.Errorf("Expected log validation status %t, given %t", desired,
				*trail.LogFileValidationEnabled)
		}

		// local state comparison
		enabled, ok := rs.Primary.Attributes["enable_log_file_validation"]
		if !ok {
			return fmt.Errorf("No enable_log_file_validation attribute defined for %s, expected %t",
				n, desired)
		}
		desiredInString := fmt.Sprintf("%t", desired)
		if enabled != desiredInString {
			return fmt.Errorf("Expected log validation status %s, saved %s", desiredInString, enabled)
		}

		return nil
	}
}

func testAccCheckCloudTrailKmsKeyIdEquals(n string, desired string, trail *cloudtrail.Trail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if desired != "" && trail.KmsKeyId == nil {
			return fmt.Errorf("No KmsKeyId attribute present in trail: %s, expected %s",
				trail, desired)
		}

		// work around string pointer
		var kmsKeyIdInString string
		if trail.KmsKeyId == nil {
			kmsKeyIdInString = ""
		} else {
			kmsKeyIdInString = *trail.KmsKeyId
		}

		if kmsKeyIdInString != desired {
			return fmt.Errorf("Expected KMS Key ID %q to equal %q",
				*trail.KmsKeyId, desired)
		}

		kmsKeyId, ok := rs.Primary.Attributes["kms_key_id"]
		if desired != "" && !ok {
			return fmt.Errorf("No kms_key_id attribute defined for %s", n)
		}
		if kmsKeyId != desired {
			return fmt.Errorf("Expected KMS Key ID %q, saved %q", desired, kmsKeyId)
		}

		return nil
	}
}

func testAccCheckAWSCloudTrailDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudtrail" {
			continue
		}

		params := cloudtrail.DescribeTrailsInput{
			TrailNameList: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeTrails(&params)

		if err == nil {
			if len(resp.TrailList) != 0 &&
				*resp.TrailList[0].Name == rs.Primary.ID {
				return fmt.Errorf("CloudTrail still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckCloudTrailLoadTags(trail *cloudtrail.Trail, tags *[]*cloudtrail.Tag) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudtrailconn
		input := cloudtrail.ListTagsInput{
			ResourceIdList: []*string{trail.TrailARN},
		}
		out, err := conn.ListTags(&input)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] Received CloudTrail tags during test: %s", out)
		if len(out.ResourceTagList) > 0 {
			*tags = out.ResourceTagList[0].TagsList
		}
		log.Printf("[DEBUG] Loading CloudTrail tags into a var: %s", *tags)
		return nil
	}
}

func testAccAWSCloudTrailConfig(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
    name = "tf-trail-foobar-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfigModified(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
    name = "tf-trail-foobar-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"
    s3_key_prefix = "/prefix"
    include_global_service_events = false
    enable_logging = false
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfigCloudWatch(randInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name = "tf-acc-test-%d"
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.test.arn}"
  cloud_watch_logs_role_arn = "${aws_iam_role.test.arn}"
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-test-trail-%d"
  force_destroy = true
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:aws:s3:::tf-test-trail-%d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::tf-test-trail-%d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name = "tf-acc-test-cloudtrail-%d"
}

resource "aws_iam_role" "test" {
  name = "tf-acc-test-cloudtrail-%d"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = "tf-acc-test-cloudtrail-%d"
  role = "${aws_iam_role.test.id}"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailCreateLogStream",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "${aws_cloudwatch_log_group.test.arn}"
    }
  ]
}
POLICY
}
`, randInt, randInt, randInt, randInt, randInt, randInt, randInt)
}

func testAccAWSCloudTrailConfigCloudWatchModified(randInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "test" {
  name = "tf-acc-test-%d"
  s3_bucket_name = "${aws_s3_bucket.test.id}"

  cloud_watch_logs_group_arn = "${aws_cloudwatch_log_group.second.arn}"
  cloud_watch_logs_role_arn = "${aws_iam_role.test.arn}"
}

resource "aws_s3_bucket" "test" {
  bucket = "tf-test-trail-%d"
  force_destroy = true
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:aws:s3:::tf-test-trail-%d"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::tf-test-trail-%d/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name = "tf-acc-test-cloudtrail-%d"
}

resource "aws_cloudwatch_log_group" "second" {
  name = "tf-acc-test-cloudtrail-second-%d"
}

resource "aws_iam_role" "test" {
  name = "tf-acc-test-cloudtrail-%d"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = "tf-acc-test-cloudtrail-%d"
  role = "${aws_iam_role.test.id}"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailCreateLogStream",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "${aws_cloudwatch_log_group.second.arn}"
    }
  ]
}
POLICY
}

`, randInt, randInt, randInt, randInt, randInt, randInt, randInt, randInt)
}

func testAccAWSCloudTrailConfigMultiRegion(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
    name = "tf-trail-foobar-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"
    is_multi_region_trail = true
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfigOrganization(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["cloudtrail.amazonaws.com"]
}

resource "aws_cloudtrail" "foobar" {
  is_organization_trail = true
  name                  = "tf-trail-foobar-%d"
  s3_bucket_name        = "${aws_s3_bucket.foo.id}"
}

resource "aws_s3_bucket" "foo" {
  bucket        = "tf-test-trail-%d"
  force_destroy = true

  policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_logValidation(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
    name = "tf-acc-trail-log-validation-test-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"
    is_multi_region_trail = true
    include_global_service_events = true
    enable_log_file_validation = true
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_logValidationModified(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
    name = "tf-acc-trail-log-validation-test-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"
    include_global_service_events = true
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_kmsKey(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %d"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_cloudtrail" "foobar" {
  name = "tf-acc-trail-log-validation-test-%d"
  s3_bucket_name = "${aws_s3_bucket.foo.id}"
  include_global_service_events = true
  kms_key_id = "${aws_kms_key.foo.arn}"
}

resource "aws_s3_bucket" "foo" {
  bucket = "tf-test-trail-%d"
  force_destroy = true
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
   {
     "Sid": "AWSCloudTrailAclCheck",
     "Effect": "Allow",
     "Principal": "*",
     "Action": "s3:GetBucketAcl",
     "Resource": "arn:aws:s3:::tf-test-trail-%d"
   },
   {
     "Sid": "AWSCloudTrailWrite",
     "Effect": "Allow",
     "Principal": "*",
     "Action": "s3:PutObject",
     "Resource": "arn:aws:s3:::tf-test-trail-%d/*",
     "Condition": {
       "StringEquals": {
         "s3:x-amz-acl": "bucket-owner-full-control"
       }
     }
   }
  ]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

var testAccAWSCloudTrailConfig_tags_tpl = `
resource "aws_cloudtrail" "foobar" {
    name = "tf-acc-trail-log-validation-test-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"
    %s
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`

func testAccAWSCloudTrailConfig_include_global_service_events(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
  name = "tf-trail-foobar-%d"
  s3_bucket_name = "${aws_s3_bucket.foo.id}"
  include_global_service_events = false
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_tags(cloudTrailRandInt int) string {
	tagsString := `tags = {
		Foo = "moo"
		Pooh = "hi"
	}`
	return fmt.Sprintf(testAccAWSCloudTrailConfig_tags_tpl,
		cloudTrailRandInt,
		tagsString,
		cloudTrailRandInt,
		cloudTrailRandInt,
		cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_tagsModified(cloudTrailRandInt int) string {
	tagsString := `tags = {
		Foo = "moo"
		Pooh = "hi"
		Moo = "boom"
	}`
	return fmt.Sprintf(testAccAWSCloudTrailConfig_tags_tpl,
		cloudTrailRandInt,
		tagsString,
		cloudTrailRandInt,
		cloudTrailRandInt,
		cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_tagsModifiedAgain(cloudTrailRandInt int) string {
	return fmt.Sprintf(testAccAWSCloudTrailConfig_tags_tpl,
		cloudTrailRandInt, "", cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_eventSelector(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
    name = "tf-trail-foobar-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"

	event_selector {
		read_write_type = "ReadOnly"
		include_management_events = false

		data_resource {
			type = "AWS::S3::Object"
			values = [
				"${aws_s3_bucket.bar.arn}/foobar",
				"${aws_s3_bucket.bar.arn}/baz",
			]
		}
	}
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}

resource "aws_s3_bucket" "bar" {
	bucket = "tf-test-trail-event-select-%d"
	force_destroy = true
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}

func testAccAWSCloudTrailConfig_eventSelectorModified(cloudTrailRandInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudtrail" "foobar" {
    name = "tf-trail-foobar-%d"
    s3_bucket_name = "${aws_s3_bucket.foo.id}"

	event_selector {
		read_write_type = "ReadOnly"
		include_management_events = true

		data_resource {
			type = "AWS::S3::Object"
			values = [
				"${aws_s3_bucket.bar.arn}/foobar",
				"${aws_s3_bucket.bar.arn}/baz",
			]
		}
	}

	event_selector {
		read_write_type = "All"
		include_management_events = false

		data_resource {
			type = "AWS::S3::Object"
			values = [
				"${aws_s3_bucket.bar.arn}/tf1",
				"${aws_s3_bucket.bar.arn}/tf2",
			]
		}

		data_resource {
			type = "AWS::Lambda::Function"
			values = [
				"${aws_lambda_function.lambda_function_test.arn}",
			]
		}
	}
}

resource "aws_s3_bucket" "foo" {
	bucket = "tf-test-trail-%d"
	force_destroy = true
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "AWSCloudTrailAclCheck",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetBucketAcl",
			"Resource": "arn:aws:s3:::tf-test-trail-%d"
		},
		{
			"Sid": "AWSCloudTrailWrite",
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:PutObject",
			"Resource": "arn:aws:s3:::tf-test-trail-%d/*",
			"Condition": {
				"StringEquals": {
					"s3:x-amz-acl": "bucket-owner-full-control"
				}
			}
		}
	]
}
POLICY
}

resource "aws_s3_bucket" "bar" {
	bucket = "tf-test-trail-event-select-%d"
	force_destroy = true
}

resource "aws_iam_role" "iam_for_lambda" {
    name = "tf-test-trail-event-select-%d"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_lambda_function" "lambda_function_test" {
    filename = "test-fixtures/lambdatest.zip"
    function_name = "tf-test-trail-event-select-%d"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs8.10"
}
`, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt, cloudTrailRandInt)
}
