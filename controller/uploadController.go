package controller

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chai2010/webp"
)

const (
	UploadImageDir  = "./tmp/image"
	UploadFileDir   = "./tmp/file"
	DefaultFileMode = 0755
	ThumbWith       = 150
	ThumbHeight     = 150
	MedWidth        = 300
	MedHeight       = 300
	BigWidth        = 1024
	BigHeight       = 1024
	QualityImage    = 80
	MaxFileSize     = 2 * 1024 * 1024 // 2MB
)

func generateUniqueFileName(oriName string) string {
	ext := filepath.Ext(oriName)
	name := strings.TrimSuffix(oriName, ext)

	// Remove special characters dari nama file
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, name)

	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}

func directoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func createDirectory(path string) error {
	if err := os.MkdirAll(path, DefaultFileMode); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func decodeImage(file io.Reader, contentType string) (image.Image, error) {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return jpeg.Decode(file)
	case "image/png":
		return png.Decode(file)
	case "image/webp":
		return webp.Decode(file)
	default:
		img, _, err := image.Decode(file)
		return img, err
	}
}

func getImageInfo(img image.Image) (width, height int) {
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func convertAndResizeImage(img image.Image, width, height uint, quality float32) ([]byte, error) {
	var resizeImage image.Image

	var buf bytes.Buffer
	err := webp.Encode(&buf, resizeImage, &webp.Options{
		Lossless: false,
		Quality:  quality,
	})

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func imageProcessing(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	defer file.Close()

	//check directory images if not exsists create directory
	thumbImageDir := UploadImageDir + "/thumb"
	if !directoryExists(thumbImageDir) {
		if err := createDirectory(thumbImageDir); err != nil {
			return "", err
		}
	}
	midImageDir := UploadImageDir + "/mid"
	if !directoryExists(midImageDir) {
		if err := createDirectory(midImageDir); err != nil {
			return "", err
		}
	}
	bigImageDir := UploadImageDir + "/big"
	if !directoryExists(bigImageDir) {
		if err := createDirectory(bigImageDir); err != nil {
			return "", err
		}
	}

	//decode image
	img, err := decodeImage(file, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	//convert dan risize file
	imgThumb, err := convertAndResizeImage(img, ThumbWith, ThumbHeight, QualityImage)
	if err != nil {
		return "", fmt.Errorf("gagal konversi ke WebP: %v", err)
	}
	imgMed, err := convertAndResizeImage(img, MedWidth, MedHeight, QualityImage)
	if err != nil {
		return "", fmt.Errorf("gagal konversi ke WebP: %v", err)
	}
	imgBig, err := convertAndResizeImage(img, BigWidth, BigHeight, QualityImage)
	if err != nil {
		return "", fmt.Errorf("gagal konversi ke WebP: %v", err)
	}

	filename := generateUniqueFileName(fileHeader.Filename)
	filepathThumb := filepath.Join(thumbImageDir, filename)
	filepathMed := filepath.Join(midImageDir, filename)
	filepathBig := filepath.Join(bigImageDir, filename)

	//simpan file
	if err := os.WriteFile(filepathThumb, imgThumb, 0755); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}
	if err := os.WriteFile(filepathMed, imgMed, 0755); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}
	if err := os.WriteFile(filepathBig, imgBig, 0755); err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}

	return filename, nil
}

func uploadController(fileHeader *multipart.FileHeader, jenis string) (string, error) {
	var namaFile string
	var err error
	switch jenis {
	case "image":
		namaFile, err = imageProcessing(fileHeader)
	case "document":

	default:
		namaFile = ""
		err = fmt.Errorf("Tidak ada Aksi yang di lakukan")
	}

	return namaFile, err
}
