package image

import (
	"image"
	"testing"
)

func TestProcessImage_OutputHasCorrectDimensions(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 2000, 2000))
	out := make(chan *image.NRGBA, 1)

	ProcessImage(src, src.Bounds(), 920, 430, out)

	result := <-out
	if result.Bounds().Dx() != 920 || result.Bounds().Dy() != 430 {
		t.Errorf("got %dx%d, want 920x430", result.Bounds().Dx(), result.Bounds().Dy())
	}
}
