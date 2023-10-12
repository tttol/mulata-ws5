package aws

import (
	"bytes"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
)

func TestUploadToS3(t *testing.T) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String("ap-northeast-1"),
		},
	}))
	bucket := "your-bucket-name" // 実際のバケット名に変更する
	key := "your-object-key"     // 実際のオブジェクトキーに変更する
	data := []byte("test-data")

	err := uploadToS3(sess, bucket, key, data)
	if err != nil {
		t.Errorf("uploadToS3 failed with error: %v", err)
	}

	// Verify that the object was uploaded to S3
	s3Svc := s3.New(sess)
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	getOutput, err := s3Svc.GetObject(getInput)
	if err != nil {
		t.Errorf("GetObject failed with error: %v", err)
	}
	defer getOutput.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(getOutput.Body)
	if buf.String() != string(data) {
		t.Errorf("GetObject returned unexpected data: %s", buf.String())
	}
}
func TestStartPolling(t *testing.T) {
	transcribeSvc := transcribeservice.New(session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})))
	jobName := "test-job"

	resultChan := make(chan string)
	errorChan := make(chan error)

	go startPolling(transcribeSvc, jobName, resultChan, errorChan)

	// Wait for the result or error
	select {
	case result := <-resultChan:
		if result == "" {
			t.Errorf("startPolling returned empty result")
		}
	case err := <-errorChan:
		t.Errorf("startPolling failed with error: %v", err)
	}
}
