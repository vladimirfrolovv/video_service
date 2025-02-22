package handlers

import (
	"context"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"

	"github.com/minio/minio-go/v7"

	"github.com/vladimirfrolovv/video-service/internal/config"
)

func UploadHandler(minioClient *minio.Client, minioCfg config.MinioConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ограничение на размер загружаемого файла — 100 МБ
		if err := r.ParseMultipartForm(100 << 20); err != nil {
			http.Error(w, "Error reading multipart form "+err.Error(), http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Dont get file from query: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		uploadInfo, err := uploadToMinIO(r.Context(), minioClient, minioCfg.BucketName, file, handler)
		if err != nil {
			log.Printf("Error get file: %v\n", err)
			http.Error(w, "Error upload file to minio: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "File download succes: %s\n", uploadInfo.ETag)
	}
}

func uploadToMinIO(ctx context.Context, client *minio.Client, bucketName string, file multipart.File, handler *multipart.FileHeader) (minio.UploadInfo, error) {
	objectName := handler.Filename

	if _, err := file.Seek(0, 0); err != nil {
		return minio.UploadInfo{}, err
	}
	return client.PutObject(
		ctx,
		bucketName,
		objectName,
		file,
		handler.Size,
		minio.PutObjectOptions{
			// 50 mb
			PartSize:    50 * pow(1024, 2),
			NumThreads:  4,
			ContentType: handler.Header.Get("Content-Type"),
		},
	)
}
func pow(a, b uint64) uint64 {
	return uint64(math.Pow(float64(a), float64(b)))
}
