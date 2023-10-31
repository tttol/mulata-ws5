package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetResult() (string, error) {
	// Set up AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	})
	if err != nil {
		fmt.Println("Error creating session:", err)
		return "", err
	}

	// Create S3 service client
	svc := s3.New(sess)

	// Set bucket and prefix
	bucket := "mulata-translate"
	prefix := "out/"

	// List objects in bucket with specified prefix
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		fmt.Println("Error listing objects:", err)
		return "", err
	}

	// Find object with latest creation date
	var latestObject *s3.Object
	var latestTime time.Time
	for _, obj := range resp.Contents {
		if obj.LastModified.After(latestTime) {
			latestObject = obj
			latestTime = *obj.LastModified
		}
	}

	// Print latest object key and creation date
	fmt.Println("Latest object:", *latestObject.Key)
	fmt.Println("Creation date:", latestTime)
	return *latestObject.Key, nil
}
