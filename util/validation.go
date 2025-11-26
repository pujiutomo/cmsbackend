package util

import (
	"fmt"
	"mime/multipart"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	MaxFileSize = 2 * 1024 * 1024 // 2MB
)

/*-----Menyimpan informasi error per field--------*/
type ValidationError struct {
	Field   string `json:"field"`   // Nama field yang error
	Message string `json:"message"` // Pesan error
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

//output {"field": "email", "message": "is required"}
/*---------------------------------------------------*/

/*---------------------Menampung kumpulan error validasi*/
type Validator struct {
	errors []ValidationError // Slice untuk menampung semua error
}

func NewValidator() *Validator {
	return &Validator{errors: make([]ValidationError, 0)}
}

/*--------------------------------------------------------*/

func (v *Validator) Required(value, fieldName string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "is required",
		})
	}
	return v
}

func (v *Validator) MinLength(value string, min int, fieldName string) *Validator {
	if len(value) < min {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be at least %d character", min),
		})
	}
	return v
}

func (v *Validator) Email(value, fieldName string) *Validator {
	if value != "" {
		//`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`
		if !regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`).MatchString(value) {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: "must be valid email address",
			})
		}
	}
	return v
}

func (v *Validator) Domain(value, fieldName string) *Validator {
	if value != "" {
		if !regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`).MatchString(value) {
			v.errors = append(v.errors, ValidationError{
				Field:   fieldName,
				Message: "must be valid domain names",
			})
		}
	}
	return v
}

func (v *Validator) FileSize(file *multipart.FileHeader, fieldName string) *Validator {
	_, err := file.Open()
	if err != nil {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "failed to get file",
		})
	}

	if file.Size > MaxFileSize {
		messageError := fmt.Sprintf("file too large. Max size is %dMB", MaxFileSize)
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: messageError,
		})
	}
	return v
}

func (v *Validator) TypeImage(file *multipart.FileHeader, fieldName string) *Validator {
	_, err := file.Open()
	if err != nil {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "failed to get file",
		})
	}
	//validasi image type
	contentType := file.Header.Get("Content-Type")
	allowedImageTypes := map[string]bool{
		"image/jpeg":               true,
		"image/png":                true,
		"image/gif":                true,
		"image/webp":               true,
		"image/bmp":                true,
		"image/x-icon":             true,
		"image/vnd.microsoft.icon": true,
	}
	if !allowedImageTypes[contentType] {
		messageError := fmt.Sprintf("file type not allowed: %s. Only images are allowed", contentType)
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: messageError,
		})
	}

	//validasi extension
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".bmp": true, ".ico": true,
	}
	if !allowedExts[ext] {
		messageError := fmt.Sprintf("file extension not allowed: %s", ext)
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: messageError,
		})
	}
	return v
}

func (v *Validator) TypeDocument(c *fiber.Ctx, fieldName string) *Validator {
	file, err := c.FormFile(fieldName)
	if err != nil {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "failed to get file",
		})
	}

	// Validasi document types
	contentType := file.Header.Get("Content-Type")
	allowedDocTypes := map[string]bool{
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
		"text/plain":                   true,
		"application/zip":              true,
		"application/x-rar-compressed": true,
	}

	if !allowedDocTypes[contentType] {
		messageError := fmt.Sprintf("document type not allowed: %s", contentType)
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: messageError,
		})
	}

	// Validasi extension
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true,
		".xls": true, ".xlsx": true, ".txt": true,
		".zip": true, ".rar": true,
	}

	if !allowedExts[ext] {
		messageError := fmt.Sprintf("document extension not allowed: %s", ext)
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: messageError,
		})
	}
	return v
}

func (v *Validator) Errors() []ValidationError {
	return v.errors //get semua error
}

func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0 //cek jika ada error
}
