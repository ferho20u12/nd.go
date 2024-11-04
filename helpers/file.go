package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	providerContract "framework-back/nucleo-de-diagnostico/contracts/providers"

	"github.com/gookit/color"
	"golang.org/x/image/draw"
)

type File struct {
	fp     providerContract.File
	config *FileConfig
}

type FileConfig struct {
	Path               string
	BaseUrl            string
	Folder             string
	DefaultCompression CompressionFile
}

type CompressionFile func(fileReader io.Reader) (io.Reader, string, error)

var (
	ErrFileContractNotDeclared = errors.New("file contract not declared")  // HTTP 500 Internal Server Error
	ErrFileTestFailed          = errors.New("error testing file provider") // HTTP 500 Internal Server Error
	ErrFileOpenFailed          = errors.New("failed to open file")         // HTTP 500 Internal Server Error
	ErrFileCompressionFailed   = errors.New("file compression failed")     // HTTP 500 Internal Server Error
	ErrFileSaveFailed          = errors.New("failed to save file")         // HTTP 500 Internal Server Error
	ErrFileReadFailed          = errors.New("error reading file")          // HTTP 500 Internal Server Error
	ErrFileEncodeImage         = errors.New("error encoding image")        // HTTP 500 Internal Server Error
	ErrFileDecodeImage         = errors.New("error decoding image")        // HTTP 500 Internal Server Error
	ErrFileEncodeImagePNG      = errors.New("error encoding image to PNG") // HTTP 500 Internal Server Error
)

func NewFile(fp providerContract.File, config FileConfig) *File {
	if fp == nil {
		color.Redln(ErrFileContractNotDeclared)
		os.Exit(1)
	}
	if err := fp.Test(); err != nil {
		color.Redf("%s: %v\n", ErrFileTestFailed, err)
		os.Exit(1)
	}
	return &File{
		fp:     fp,
		config: &config,
	}
}

func (f *File) ValidateFolder(folder string) bool {
	return f.config.Folder == folder
}

func (f *File) GenerateUrl(fileName string) string {
	return fmt.Sprintf("%s/?folder=%s&fileName=%s", f.config.BaseUrl, f.config.Folder, fileName)
}

func (f *File) SaveFile(file *multipart.FileHeader, fileName string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFileOpenFailed, err)
	}
	defer src.Close()

	var fileReader io.Reader = src
	if f.config.DefaultCompression != nil {
		compressedFileReader, newExt, err := f.config.DefaultCompression(src)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrFileCompressionFailed, err)
		}
		fileReader = compressedFileReader
		if newExt != "" {
			ext = newExt
		}
	}
	path := fmt.Sprintf("%s/%s", f.config.Path, fileName+ext)
	if err := f.fp.SaveFile(fileReader, path); err != nil {
		return "", fmt.Errorf("%w: %v", ErrFileSaveFailed, err)
	}

	return fileName + ext, nil
}

func (f *File) Read(fileName string) (io.ReadCloser, error) {
	path := fmt.Sprintf("%s/%s", f.config.Path, fileName)
	return f.fp.ReadFile(path)
}

func (f *File) Delete(fileName string) error {
	path := fmt.Sprintf("%s/%s", f.config.Path, fileName)
	return f.fp.DeleteFile(path)
}

// DefaultCompressImageToJPG compresses and resizes an image to the specified dimensions if it exceeds the limits.
// Returns: (io.Reader, string, error)
func DefaultCompressImageToJPG(fileReader io.Reader) (io.Reader, string, error) {
	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, fileReader); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileReadFailed, err)
	}
	img, _, err := image.Decode(&buffer)
	if err != nil {
		img, err = png.Decode(fileReader)
		if err != nil {
			return &buffer, "", nil
		}
	}
	// Check the image size and resize only if necessary
	maxWidth, maxHeight := 800, 600
	img = ResizeImage(img, maxWidth, maxHeight)

	var compressed bytes.Buffer
	opts := jpeg.Options{Quality: 80}
	if err := jpeg.Encode(&compressed, img, &opts); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileEncodeImage, err)
	}
	return &compressed, ".jpg", nil
}

// DefaultCompressImageToPNG compresses and resizes an image to the specified dimensions if it exceeds the limits.
// Returns: (io.Reader, string, error)
func DefaultCompressImageToPNG(fileReader io.Reader) (io.Reader, string, error) {
	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, fileReader); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileReadFailed, err)
	}

	img, _, err := image.Decode(&buffer)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileDecodeImage, err)
	}

	// Check the image size and resize only if necessary
	maxWidth, maxHeight := 640, 480
	img = ResizeImage(img, maxWidth, maxHeight)

	// Convert to PNG
	var compressed bytes.Buffer
	if err := png.Encode(&compressed, img); err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrFileEncodeImagePNG, err)
	}

	return &compressed, ".png", nil
}

// ResizeImage resizes an image to the specified maximum width and height while maintaining the aspect ratio.
// Returns: image.Image
func ResizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= maxWidth && height <= maxHeight {
		return img
	}
	var newWidth, newHeight int
	aspectRatio := float64(width) / float64(height)

	if aspectRatio > 1 {
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / aspectRatio)
		if newHeight > maxHeight {
			newHeight = maxHeight
			newWidth = int(float64(maxHeight) * aspectRatio)
		}
	} else {
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * aspectRatio)
		if newWidth > maxWidth {
			newWidth = maxWidth
			newHeight = int(float64(maxWidth) / aspectRatio)
		}
	}
	resizedImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(resizedImg, resizedImg.Bounds(), img, bounds, draw.Over, nil)
	return resizedImg
}
