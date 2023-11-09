package aws

import (
	"io/ioutil"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func GetResult() (string, error) {
	sssion := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))
	svc := s3.New(sssion)

	bucket := "mulata-translate"
	prefix := "out/"
	resp, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		slog.Error("Error listing objects:", err)
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
	slog.Info("Latest object:", latestObject.Key)

	// Get the object
	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    latestObject.Key,
	})
	if err != nil {
		slog.Error("Error getting object:", err)
		return "", err
	}
	defer obj.Body.Close()

	// Read the object's content
	body, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		slog.Error("Error reading object body:", err)
		return "", err
	}

	// Convert the body to a string and return it
	slog.Info("Success to get result:", string(body))
	return string(body), nil
}
