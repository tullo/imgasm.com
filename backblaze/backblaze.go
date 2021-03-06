package backblaze

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/packago/config"
	"github.com/tullo/imgasm/models"
)

// Upload a file to a bucket in backblaze cloude storrage.
func Upload(log *log.Logger, file models.File) error {
	acc, err := authorizeAccount()
	if err != nil {
		return fmt.Errorf("failed to authorize %v", err)
	}

	key := aws.String(fmt.Sprintf("%x.%s", file.MD5Hash, file.Extension))
	meta := make(map[string]*string)
	meta["x-bz-info-author"] = aws.String("unknown")
	meta["x-bz-file-name"] = aws.String(fmt.Sprintf("%x.%s", file.MD5Hash, file.Extension))
	meta["x-bz-content-sha1"] = aws.String(fmt.Sprintf("%x", sha1.Sum(file.Body)))
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.File().GetString("backblaze.keyID"),
			config.File().GetString("backblaze.applicationKey"),
			config.File().GetString("backblaze.token"),
		),
		Endpoint:         aws.String(config.File().GetString("backblaze.s3.endpoint")),
		Region:           aws.String("eu-central-003"),
		S3ForcePathStyle: aws.Bool(true),
	}
	s, err := session.NewSession(s3Config)
	if err != nil {
		return fmt.Errorf("failed to upload object %s/%s, %v", acc.Allowed.BucketName, *key, err)
	}
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Body:        bytes.NewReader(file.Body),
		Bucket:      &acc.Allowed.BucketName,
		Key:         key,
		ContentType: &file.MimeType,
		Metadata:    meta,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %s/%s, %v", acc.Allowed.BucketName, *key, err)
	}
	log.Printf("Successfully uploaded file: %s\n", *key)
	return nil
}
