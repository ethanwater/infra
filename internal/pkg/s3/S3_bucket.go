package s3

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func FetchBucketObjects() ([]*s3.Object, error) {
	init_config()
	sess, err := session.NewSession(&aws.Config{Region: aws.String((os.Getenv("AWS_REGION")))})
	if err != nil {
		return nil, err
	}
	svc := s3.New(sess)
	input := &s3.ListObjectsV2Input{Bucket: aws.String(os.Getenv("AWS_BUCKET"))}

	resp, err := svc.ListObjectsV2(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
			return nil, err
		}
	}

	return resp.Contents, nil
}
