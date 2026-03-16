package storage

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(endpoint, accessKey, secretKey, bucketName string) (*S3Storage, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // У нас локально нет HTTPS
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	
	// Проверяем, существует ли бакет, если нет - создаем
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		log.Printf("Bucket '%s' successfully created\n", bucketName)
	}

	log.Println("Connected to MinIO (S3) successfully")
	return &S3Storage{client: minioClient, bucket: bucketName}, nil
}

// UploadImage загружает файл в S3 и возвращает ссылку на него
func (s *S3Storage) UploadImage(ctx context.Context, objectName string, filePath string, contentType string) (string, error) {
	_, err := s.client.FPutObject(ctx, s.bucket, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	
	// Возвращаем публичный URL картинки
	url := "http://localhost:9000/" + s.bucket + "/" + objectName
	return url, nil
}
