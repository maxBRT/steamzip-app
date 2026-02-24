package image

import (
	"github.com/disintegration/imaging"
	"image"
)

func ProcessImage(masterImage image.Image, rect image.Rectangle, targetW, targetH int, out chan<- *image.NRGBA) {
	cropped := imaging.Crop(masterImage, rect)
	resized := imaging.Resize(cropped, targetW, targetH, imaging.Lanczos)
	out <- resized
}
