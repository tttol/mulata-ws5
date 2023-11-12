package aws

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/transcribeservice"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO originチェックちゃんとする
		return true
	},
}

func toMp3(inputFile string) ([]byte, error) {
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-f", "mp3", "-")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error encoding audio: %v (stderr: %s)", err, stderr.String())
		// return nil, fmt.Errorf("error encoding audio: %v", err)
	}
	return out.Bytes(), nil
}

func uploadToS3(sess *session.Session, bucket, key string, data []byte) error {
	s3Svc := s3.New(sess)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}
	_, err := s3Svc.PutObject(input)
	return err
}

func startPolling(transcribeSvc *transcribeservice.TranscribeService, jobName string, resultChan chan string, errorChan chan error) {
	for {
		slog.Info("Start polling...")
		time.Sleep(5 * time.Second) // Wait before checking the job status

		jobInput := &transcribeservice.GetTranscriptionJobInput{
			TranscriptionJobName: aws.String(jobName),
		}
		jobOutput, err := transcribeSvc.GetTranscriptionJob(jobInput)
		if err != nil {
			errorChan <- fmt.Errorf("error getting transcription job: %v", err)
			return
		}

		if *jobOutput.TranscriptionJob.TranscriptionJobStatus == "COMPLETED" {
			resultChan <- *jobOutput.TranscriptionJob.Transcript.TranscriptFileUri
			return
		} else if *jobOutput.TranscriptionJob.TranscriptionJobStatus == "FAILED" {
			errorChan <- fmt.Errorf("transcription job failed. Reason: %s", *jobOutput.TranscriptionJob.FailureReason)
			return
		}
		slog.Info("End polling...")
	}
}

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Error during connection upgrade:", err)
		return
	}
	defer ws.Close()

	slog.Info("Client Connected!")

	// Initialize AWS session
	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))

	// Initialize AWS Transcribe service client
	transcribeSvc := transcribeservice.New(session)

	for {
		_, audioData, err := ws.ReadMessage()
		if err != nil {
			slog.Error("Error during reading message:", err)
			break
		}

		// msgを一時ファイルに書き込む
		tmpFile, err := ioutil.TempFile("", "audio-*.webm")
		if err != nil {
			slog.Error("Error creating temporary file:", err)
			continue
		}
		if _, err := tmpFile.Write(audioData); err != nil {
			slog.Error("Error writing to temporary file:", err)
			continue
		}

		// ffmpegを使用して、一時ファイルをMP3ファイルに変換する
		out, err := toMp3(tmpFile.Name())
		if err != nil {
			slog.Error("Error encoding audio:", "err", err)
			continue
		}
		slog.Info("Delete temporary file:", tmpFile.Name())
		os.Remove(tmpFile.Name())

		// mp3ファイルをS3にアップロードする
		bucket := "mulata-appfile"
		mp3Key := fmt.Sprintf("audio/audio_%s.mp3", time.Now().Format("20231231150405.000"))
		if err := uploadToS3(session, bucket, mp3Key, out); err != nil {
			slog.Error("Error uploading to S3:", err)
			continue
		}
		slog.Info("Uploading to S3 has succeeded.", "bucket", bucket, "key", mp3Key)
		s3Uri := fmt.Sprintf("s3://%s/%s", bucket, mp3Key)

		// Start transcription job
		jobName := fmt.Sprintf("TranscriptionJob_%d", time.Now().Unix())
		outputKey := fmt.Sprintf("transcribe/out/transcribe_%s.json", time.Now().Format("20231231150405.000"))
		input := &transcribeservice.StartTranscriptionJobInput{
			LanguageCode: aws.String("en-US"),
			Media: &transcribeservice.Media{
				MediaFileUri: aws.String(s3Uri),
			},
			MediaFormat:          aws.String("mp3"),
			TranscriptionJobName: aws.String(jobName),
			OutputBucketName:     aws.String(bucket),
			OutputKey:            aws.String(outputKey),
		}

		_, err = transcribeSvc.StartTranscriptionJob(input)
		if err != nil {
			slog.Error("Error starting transcription job:", err)
			continue
		}

		resultChan := make(chan string)
		errorChan := make(chan error)

		go startPolling(transcribeSvc, jobName, resultChan, errorChan)

		select {
		case result := <-resultChan:
			slog.Info("Transcription completed. Result URL:%s", result)
		case err := <-errorChan:
			slog.Error("Error during transcription:", err)
		case <-time.After(1 * time.Minute):
			slog.Error("Timed out waiting for transcription result")
		}
	}
}

func Get() (string, error) {
	sssion := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))
	svc := s3.New(sssion)

	bucket := "mulata-appfile"
	prefix := "transcribe/out/"
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
