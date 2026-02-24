package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	steamimage "github.com/maxbrt/steamzip-app/internal/image"
	"github.com/maxbrt/steamzip-app/internal/models"
)

func (h *Handler) runProcessingJob(asset models.Asset, focalPoints map[string]Point) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic in processing job for asset %s: %v", asset.ID, rec)
			if dbErr := h.Database.Model(&asset).Update("status", models.AssetStatusFailed).Error; dbErr != nil {
				log.Printf("failed to mark asset %s as failed: %v", asset.ID, dbErr)
			}
		}
	}()

	ctx := context.Background()

	fail := func(reason string, err error) {
		log.Printf("processing job failed for asset %s: %s: %v", asset.ID, reason, err)
		if dbErr := h.Database.Model(&asset).Update("status", models.AssetStatusFailed).Error; dbErr != nil {
			log.Printf("failed to mark asset %s as failed: %v", asset.ID, dbErr)
		}
	}

	// 1. Download master image from S3
	obj, err := h.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(h.S3Bucket),
		Key:    aws.String(asset.MasterFileKey),
	})
	if err != nil {
		fail("download master", err)
		return
	}
	defer obj.Body.Close()

	masterImage, err := imaging.Decode(obj.Body)
	if err != nil {
		fail("decode master image", err)
		return
	}

	// 2. Crop all 14 asset types in parallel
	type cropResult struct {
		key string
		img *image.NRGBA
	}

	results := make([]cropResult, len(steamimage.SteamAssets))
	var wg sync.WaitGroup

	for i, spec := range steamimage.SteamAssets {
		wg.Add(1)
		go func(i int, spec steamimage.SteamAssetSpec) {
			defer wg.Done()
			point := focalPoints[spec.Key] // zero value Point{0,0,0} if not provided
			rect := steamimage.CropRect(masterImage.Bounds(), point.X, point.Y, point.Zoom, spec)
			out := make(chan *image.NRGBA, 1)
			steamimage.ProcessImage(masterImage, rect, spec.Width, spec.Height, out)
			results[i] = cropResult{key: spec.Key, img: <-out}
		}(i, spec)
	}
	wg.Wait()

	// 3. Bundle cropped images into a ZIP
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, r := range results {
		fw, err := zw.Create(r.key + ".png")
		if err != nil {
			fail("create zip entry "+r.key, err)
			return
		}
		if err := png.Encode(fw, r.img); err != nil {
			fail("encode png "+r.key, err)
			return
		}
	}
	if err := zw.Close(); err != nil {
		fail("close zip", err)
		return
	}

	// 4. Upload ZIP to S3
	zipKey := fmt.Sprintf("zips/%s.zip", asset.ID)
	_, err = h.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(h.S3Bucket),
		Key:         aws.String(zipKey),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("application/zip"),
		Metadata: map[string]string{
			"session-id": asset.SessionID.String(),
			"asset-id":   asset.ID.String(),
		},
	})
	if err != nil {
		fail("upload zip", err)
		return
	}

	// 5. Mark asset as ready
	if err := h.Database.Model(&asset).Updates(map[string]any{
		"status":  models.AssetStatusReady,
		"zip_key": zipKey,
	}).Error; err != nil {
		log.Printf("failed to update asset %s after processing: %v", asset.ID, err)
	}
}
