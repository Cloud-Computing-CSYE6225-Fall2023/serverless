package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/sns"
	"os"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/serviceaccount"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type GCPCredentials struct {
	Type                string `json:"type"`
	ProjectID           string `json:"project_id"`
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ClientEmail         string `json:"client_email"`
	ClientID            string `json:"client_id"`
	AuthURI             string `json:"auth_uri"`
	TokenURI            string `json:"token_uri"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url"`
	ClientCertUrl       string `json:"client_x509_cert_url"`
	UniverseDomain      string `json:"universe_domain"`
}

type Resources struct {
	Region               string `json:"region"`
	AccountID            string `json:"account_id"`
	SNSTopicName         string `json:"sns_topic_name"`
	SQSTopicName         string `json:"sqs_topic_name"`
	LambdaLogGroup       string `json:"lambda_log_group"`
	LambdaFuncName       string `json:"lambda_func_name"`
	LambdaFilePath       string `json:"lambda_file_path"`
	DependenciesFilePath string `json:"dependencies_file_path"`
	GcpProjectID         string `json:"gcp_project_id"`
	GcpBucketName        string `json:"gcp_bucket_name"`
	DynamoDbTableName    string `json:"dynamo_db_table_name"`
	MailgunDomainName    string `json:"mailgun_domain_name"`
	MailgunApiKey        string `json:"mailgun_api_key"`
	MailgunTemplate      string `json:"mailgun_template"`
}

type Data struct {
	ResourceParams Resources `json:"resource_params"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Load configuration values from pulumi.*.yaml file
		var configData Data
		cfg := config.New(ctx, "")
		cfg.RequireObject("config", &configData)

		lambdaRoleStr, err := json.Marshal(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect": "Allow",
					"Action": []string{"sts:AssumeRole"},
					"Principal": map[string]string{
						"Service": "lambda.amazonaws.com",
					},
				},
			},
		})

		lambdaRole, err := iam.NewRole(ctx, "lambdaRole", &iam.RoleArgs{
			AssumeRolePolicy: pulumi.String(lambdaRoleStr),
			Tags: pulumi.StringMap{
				"tag-key": pulumi.String("lambda-role"),
			},
		})
		if err != nil {
			return err
		}

		lambdaExecutionPolicyStr, err := json.Marshal(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect":   "Allow",
					"Action":   "logs:CreateLogGroup",
					"Resource": fmt.Sprintf("arn:aws:logs:%v:%v:*", configData.ResourceParams.Region, configData.ResourceParams.AccountID),
				},
				{
					"Action": []string{
						"logs:CreateLogStream",
						"logs:PutLogEvents",
					},
					"Effect":   "Allow",
					"Resource": fmt.Sprintf("arn:aws:logs:%v:%v:log-group:/aws/lambda/%v:*", configData.ResourceParams.Region, configData.ResourceParams.AccountID, configData.ResourceParams.LambdaLogGroup),
				},
			},
		})
		if err != nil {
			return err
		}

		customLambdaExecutionPolicy, err := iam.NewPolicy(ctx, "CustomLambdaPolicy", &iam.PolicyArgs{
			Path:        pulumi.String("/"),
			Description: pulumi.String("Custom SNS policy to publish message from EC2 to SNS Topic"),
			Policy:      pulumi.String(lambdaExecutionPolicyStr),
		})
		if err != nil {
			return err
		}

		_, err = iam.NewRolePolicyAttachment(ctx, "ec2LamdaPolicy", &iam.RolePolicyAttachmentArgs{
			Role:      lambdaRole.ID(),
			PolicyArn: customLambdaExecutionPolicy.Arn,
		})
		if err != nil {
			return err
		}

		//lambdaPolicyStr, err := json.Marshal(map[string]interface{}{
		//	"Version": "2012-10-17",
		//	"Statement": []map[string]interface{}{
		//		{
		//			"Sid":    "VisualEditor0",
		//			"Effect": "Allow",
		//			"Action": []string{
		//				"lambda:CreateEventSourceMapping",
		//				"lambda:ListEventSourceMappings",
		//				"lambda:ListFunctions",
		//			},
		//			"Resource": "*",
		//		},
		//	},
		//})
		//if err != nil {
		//	return err
		//}

		//lambdaPolicy, err := iam.NewPolicy(ctx, "lambdaLoggingPolicy", &iam.PolicyArgs{
		//	Path:        pulumi.String("/"),
		//	Description: pulumi.String("IAM policy for logging from a lambda"),
		//	Policy:      pulumi.String(lambdaPolicyStr),
		//})
		//if err != nil {
		//	return err
		//}
		//
		//// Attach the policy to the role
		//_, err = iam.NewRolePolicyAttachment(ctx, "lambdaLogs", &iam.RolePolicyAttachmentArgs{
		//	Role:      lambdaRole.ID(),
		//	PolicyArn: lambdaPolicy.Arn,
		//})
		//if err != nil {
		//	return err
		//}

		//lambdaFunctionName := "gcp_file_storage"

		// Create a log group for the lambda function logs
		lambdaLogGroup, err := cloudwatch.NewLogGroup(ctx, "example", &cloudwatch.LogGroupArgs{
			Name:            pulumi.String(configData.ResourceParams.LambdaFuncName),
			RetentionInDays: pulumi.Int(5),
		})
		if err != nil {
			return err
		}

		// Change the file permissions
		//filePath := "./gcp_file_storage.zip"
		//dependenciesPath := "./dependencies.zip"
		permissions := os.FileMode(0644)

		err = os.Chmod(configData.ResourceParams.LambdaFilePath, permissions)
		if err != nil {
			return err
		}

		err = os.Chmod(configData.ResourceParams.DependenciesFilePath, permissions)
		if err != nil {
			return err
		}

		dependenciesLayer, err := lambda.NewLayerVersion(ctx, "lambdaLayer", &lambda.LayerVersionArgs{
			CompatibleRuntimes: pulumi.StringArray{
				pulumi.String("python3.11"),
			},
			CompatibleArchitectures: pulumi.StringArray{
				pulumi.String("arm64"),
			},
			Description: pulumi.String("This Layers contains all python runtime dependencies"),
			Code:        pulumi.NewFileArchive(configData.ResourceParams.DependenciesFilePath),
			LayerName:   pulumi.String("python_dependencies"),
		})
		if err != nil {
			return err
		}

		//
		_, err = storage.NewBucket(ctx, "submission-bucket", &storage.BucketArgs{
			Project:                  pulumi.String(configData.ResourceParams.GcpProjectID),
			Name:                     pulumi.String(configData.ResourceParams.GcpBucketName),
			ForceDestroy:             pulumi.Bool(true),
			Location:                 pulumi.String("US-EAST1"),
			StorageClass:             pulumi.String("STANDARD"),
			UniformBucketLevelAccess: pulumi.Bool(true),
			PublicAccessPrevention:   pulumi.String("enforced"),
			Versioning: storage.BucketVersioningArgs{
				Enabled: pulumi.Bool(false),
			},
			//Logging:                  storage.BucketLoggingArgs{
			//	LogBucket: pulumi.String(""),
			//	LogObjectPrefix: pulumi.String(""),
			//},
		})
		if err != nil {
			return err
		}

		// Create a Service Account
		sa, err := serviceaccount.NewAccount(ctx, "bucket-access-service-account", &serviceaccount.AccountArgs{
			AccountId:   pulumi.String("bucketaccess"),
			Project:     pulumi.String(configData.ResourceParams.GcpProjectID),
			DisplayName: pulumi.String("ObjectReadWriteAccess"),
			Disabled:    pulumi.Bool(false),
			Description: pulumi.String("Service account which has access to read and write to a bucket"),
		})
		if err != nil {
			return err
		}

		// Make the service account an objectViewer on the bucket
		_, err = storage.NewBucketIAMMember(ctx, "bucket-object-viewer", &storage.BucketIAMMemberArgs{
			Bucket: pulumi.String(configData.ResourceParams.GcpBucketName),
			Role:   pulumi.String("roles/storage.objectViewer"),
			Member: sa.Email.ApplyT(func(email string) string { return "serviceAccount:" + email }).(pulumi.StringInput),
		})
		if err != nil {
			return err
		}

		// Make the service account an objectCreator on the bucket
		_, err = storage.NewBucketIAMMember(ctx, "bucket-object-creator", &storage.BucketIAMMemberArgs{
			Bucket: pulumi.String(configData.ResourceParams.GcpBucketName),
			Role:   pulumi.String("roles/storage.objectCreator"),
			Member: sa.Email.ApplyT(func(email string) string { return "serviceAccount:" + email }).(pulumi.StringInput),
		})
		if err != nil {
			return err
		}

		//_, err = serviceaccount.NewIAMBinding(ctx, "bucket-object-viewer", &serviceaccount.IAMBindingArgs{
		//	ServiceAccountId: sa.Name,
		//	Role:             pulumi.String("roles/storage.objectViewer"),
		//	Condition: &serviceaccount.IAMBindingConditionArgs{
		//		Title:       pulumi.String("Bucket Object Viewer"),
		//		Description: pulumi.String("User can only view objects of a specified bucket"),
		//		Expression:  pulumi.String("resource.name.startsWith('projects/csye6125-dev/buckets/assignment-submission-test1')"),
		//	},
		//	Members: pulumi.StringArray{
		//		pulumi.String("allAuthenticatedUsers"),
		//	},
		//})

		// Create a key for the Service Account
		key, err := serviceaccount.NewKey(ctx, "service-account-key", &serviceaccount.KeyArgs{
			ServiceAccountId: sa.Name,
			KeyAlgorithm:     pulumi.String("KEY_ALG_RSA_2048"),
			PublicKeyType:    pulumi.String("TYPE_X509_PEM_FILE"),
			PrivateKeyType:   pulumi.String("TYPE_GOOGLE_CREDENTIALS_FILE"),
		})
		if err != nil {
			return err
		}

		key.PrivateKey.ApplyT(func(pkVal string) error {
			var gcpCreds GCPCredentials
			// Decode the base64-encoded string
			decodedBytes, err := base64.StdEncoding.DecodeString(pkVal)
			if err != nil {
				return err
			}

			// Convert the decoded bytes to a string
			credentials := string(decodedBytes)
			err = json.Unmarshal([]byte(credentials), &gcpCreds)
			if err != nil {
				return err
			}

			//
			lambdaFunc, err := lambda.NewFunction(ctx, "file-store-lambda", &lambda.FunctionArgs{
				Name:    pulumi.String(configData.ResourceParams.LambdaFuncName),
				Handler: pulumi.String("lambda_function.lambda_handler"),
				Role:    lambdaRole.Arn,
				Timeout: pulumi.Int(15),
				Runtime: pulumi.String("python3.11"),
				Architectures: pulumi.StringArray{
					pulumi.String("arm64"),
				},
				Code: pulumi.NewFileArchive(configData.ResourceParams.LambdaFilePath),
				Layers: pulumi.StringArray{
					dependenciesLayer.Arn,
				},
				Environment: &lambda.FunctionEnvironmentArgs{
					Variables: pulumi.StringMap{
						"region":                      pulumi.String(configData.ResourceParams.Region),
						"type":                        pulumi.String(gcpCreds.Type),
						"project_id":                  pulumi.String(gcpCreds.ProjectID),
						"private_key_id":              pulumi.String(gcpCreds.PrivateKeyID),
						"private_key":                 pulumi.String(gcpCreds.PrivateKey),
						"client_email":                pulumi.String(gcpCreds.ClientEmail),
						"client_id":                   pulumi.String(gcpCreds.ClientID),
						"auth_uri":                    pulumi.String(gcpCreds.AuthURI),
						"token_uri":                   pulumi.String(gcpCreds.TokenURI),
						"client_x509_cert_url":        pulumi.String(gcpCreds.ClientCertUrl),
						"universe_domain":             pulumi.String(gcpCreds.UniverseDomain),
						"auth_provider_x509_cert_url": pulumi.String(gcpCreds.AuthProviderCertURL),
						"dynamodb_table_name":         pulumi.String(configData.ResourceParams.DynamoDbTableName),
						"bucket_name":                 pulumi.String(configData.ResourceParams.GcpBucketName),
						"mailgun_domain_name":         pulumi.String(configData.ResourceParams.MailgunDomainName),
						"mailgun_api_key":             pulumi.String(configData.ResourceParams.MailgunApiKey),
						"mailgun_template":            pulumi.String(configData.ResourceParams.MailgunTemplate),
					},
				},
			}, pulumi.DependsOn([]pulumi.Resource{
				lambdaLogGroup,
			}))
			if err != nil {
				return err
			}

			// Create a new event source mapping which connects the SQS queue with the Lambda function
			//sqsArn := fmt.Sprintf("arn:aws:sqs:%v:%v:%v", configData.ResourceParams.Region, configData.ResourceParams.AccountID, configData.ResourceParams.SQSTopicName)
			//_, err = lambda.NewEventSourceMapping(ctx, "sqsEventSource", &lambda.EventSourceMappingArgs{
			//	EventSourceArn: pulumi.String(sqsArn),
			//	FunctionName:   lambdaFunc.Arn,
			//	Enabled:        pulumi.Bool(true),
			//	BatchSize:      pulumi.Int(1),
			//})
			//if err != nil {
			//	return err
			//}

			_, err = lambda.NewPermission(ctx, "withSns", &lambda.PermissionArgs{
				Action:    pulumi.String("lambda:InvokeFunction"),
				Function:  lambdaFunc.Name,
				Principal: pulumi.String("sns.amazonaws.com"),
				SourceArn: pulumi.String(fmt.Sprintf("arn:aws:sns:%v:%v:%v", configData.ResourceParams.Region, configData.ResourceParams.AccountID, configData.ResourceParams.SNSTopicName)),
			})
			if err != nil {
				return err
			}

			_, err = sns.NewTopicSubscription(ctx, "submissionsSQSTarget", &sns.TopicSubscriptionArgs{
				Topic:    pulumi.String(fmt.Sprintf("arn:aws:sns:%v:%v:%v", configData.ResourceParams.Region, configData.ResourceParams.AccountID, configData.ResourceParams.SNSTopicName)),
				Protocol: pulumi.String("lambda"),
				Endpoint: lambdaFunc.Arn,
			})
			if err != nil {
				return err
			}

			// Create DynamoDB Table
			_, err = dynamodb.NewTable(ctx, "email-tracking-table", &dynamodb.TableArgs{
				Attributes: dynamodb.TableAttributeArray{
					&dynamodb.TableAttributeArgs{
						Name: pulumi.String("id"),
						Type: pulumi.String("S"),
					},
				},
				Name:         pulumi.String(configData.ResourceParams.DynamoDbTableName),
				BillingMode:  pulumi.String("PROVISIONED"),
				HashKey:      pulumi.String("id"),
				ReadCapacity: pulumi.Int(20),
				Tags: pulumi.StringMap{
					"Name": pulumi.String(configData.ResourceParams.DynamoDbTableName),
				},
				Ttl: &dynamodb.TableTtlArgs{
					AttributeName: pulumi.String("TimeToExist"),
					Enabled:       pulumi.Bool(false),
				},
				WriteCapacity: pulumi.Int(20),
			})
			if err != nil {
				return err
			}

			lambdaDynamoDBPolicyStr, err := json.Marshal(map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect": "Allow",
						"Action": []string{
							"dynamodb:CreateTable",
							"dynamodb:UpdateTimeToLive",
							"dynamodb:PutItem",
							"dynamodb:DescribeTable",
							"dynamodb:DeleteItem",
							"dynamodb:GetItem",
							"dynamodb:Scan",
							"dynamodb:Query",
							"dynamodb:UpdateItem",
							"dynamodb:UpdateTable",
						},
						"Resource": fmt.Sprintf("arn:aws:dynamodb:%v:%v:table/%v", configData.ResourceParams.Region, configData.ResourceParams.AccountID, configData.ResourceParams.DynamoDbTableName),
					},
				},
			})
			if err != nil {
				return err
			}

			lambdaDynamoDbPolicy, err := iam.NewPolicy(ctx, "lambdaDynamoDBPolicy", &iam.PolicyArgs{
				Path:        pulumi.String("/"),
				Description: pulumi.String("IAM policy for logging from a lambda"),
				Policy:      pulumi.String(lambdaDynamoDBPolicyStr),
			})
			if err != nil {
				return err
			}

			// Attach the policy to the role
			_, err = iam.NewRolePolicyAttachment(ctx, "lambdaDynamoPolicyAttachment", &iam.RolePolicyAttachmentArgs{
				Role:      lambdaRole.ID(),
				PolicyArn: lambdaDynamoDbPolicy.Arn,
			})
			if err != nil {
				return err
			}

			return nil
		})

		return nil
	})
}
