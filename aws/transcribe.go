package aws

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	}
}

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error during connection upgrade:", err)
		return
	}
	defer ws.Close()

	fmt.Println("Client Connected!")

	// Initialize AWS session
	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))

	// Initialize AWS Transcribe service client
	transcribeSvc := transcribeservice.New(session)

	for {
		_, audioMsg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Error during reading message:", err)
			break
		}

		// msgを一時ファイルに書き込む
		tmpFile, err := ioutil.TempFile("", "audio-*.webm")
		if err != nil {
			fmt.Println("Error creating temporary file:", err)
			return
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.Write(audioMsg); err != nil {
			fmt.Println("Error writing to temporary file:", err)
			return
		}

		// ffmpegを使用して、一時ファイルをMP3ファイルに変換する
		cmd := exec.Command("ffmpeg", "-i", tmpFile.Name(), "-vn", "-acodec", "libmp3lame", "-qscale:a", "2", "-f", "mp3", "-")
		out, err := cmd.Output()
		if err != nil {
			fmt.Println("Error encoding audio:", err)
			return
		}

		// mp3ファイルをS3にアップロードする
		bucket := "mulata-appfile"
		key := fmt.Sprintf("audio/audio_%s.mp3", time.Now().Format("20231231150405.000"))
		if err := uploadToS3(session, bucket, key, out); err != nil {
			fmt.Println("Error uploading to S3:", err)
			return
		}

		s3Uri := fmt.Sprintf("s3://%s/%s", bucket, key)

		// Start transcription job
		jobName := fmt.Sprintf("TranscriptionJob_%d", time.Now().Unix())
		input := &transcribeservice.StartTranscriptionJobInput{
			LanguageCode: aws.String("en-US"),
			Media: &transcribeservice.Media{
				MediaFileUri: aws.String(s3Uri),
			},
			MediaFormat:          aws.String("mp3"),
			TranscriptionJobName: aws.String(jobName),
		}

		_, err = transcribeSvc.StartTranscriptionJob(input)
		if err != nil {
			fmt.Println("Error starting transcription job:", err)
			continue
		}

		resultChan := make(chan string)
		errorChan := make(chan error)

		go startPolling(transcribeSvc, jobName, resultChan, errorChan)

		select {
		case result := <-resultChan:
			fmt.Println("Transcription completed. Result URL:", result)
		case err := <-errorChan:
			fmt.Println("Error during transcription:", err)
		case <-time.After(1 * time.Minute):
			fmt.Println("Timed out waiting for transcription result")
		}
	}
}
