package internal

import (
	"context"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func CopyFromContainerToContainer(docker *client.Client, srcContainerID string, srcPath string, dstContainerID string, dstPath string) {
	// Create a new context
	ctx := context.Background()

	// Open a new reader for the file in the source container
	reader, _, err := docker.CopyFromContainer(ctx, srcContainerID, srcPath)
	if err != nil {
		log.Fatalf("could not copy files %s from container %s to host\n", srcPath, srcContainerID)
	}

	defer reader.Close()

	// Open a new writer for the file in the destination container
	err = docker.CopyToContainer(ctx, dstContainerID, dstPath, reader, types.CopyToContainerOptions{})
	if err != nil {
		log.Fatalf("could not copy files from container %s to container %s\n", srcContainerID, dstContainerID)
	}
}

func UploadArtifactFromContainer(docker *client.Client, pipelineName string, stageName string, srcContainerID string, srcPath string) string {
	// Set the S3 bucket and destination path
	bucket := "big-data-ci"
	dstPath := pipelineName + "/" + stageName + "/artifacts/"

	accessKey, secretKey, region := GetAWSCreds()

	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create a new S3 manager
	manager := s3manager.NewUploader(sess)

	// Open a new reader for the file in the container
	reader, _, err := docker.CopyFromContainer(context.Background(), srcContainerID, srcPath)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	// Create an S3 upload input
	input := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstPath),
		Body:   reader,
	}

	// Upload the file to S3
	result, err := manager.Upload(input)
	if err != nil {
		log.Fatal(err)
	}

	return result.Location
}
