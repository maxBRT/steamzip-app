package image

import (
	"image"
	"math"
)

type SteamAssetSpec struct {
	Key    string
	Width  int
	Height int
}

var SteamAssets = []SteamAssetSpec{
	{Key: "header_capsule", Width: 920, Height: 430},
	{Key: "small_capsule", Width: 462, Height: 174},
	{Key: "main_capsule", Width: 1232, Height: 706},
	{Key: "vertical_capsule", Width: 748, Height: 896},
	{Key: "screenshots", Width: 1920, Height: 1080},
	{Key: "page_background", Width: 1438, Height: 810},
	{Key: "shortcut_icon", Width: 256, Height: 256},
	{Key: "app_icon", Width: 184, Height: 184},
	{Key: "library_capsule", Width: 600, Height: 900},
	{Key: "library_hero", Width: 3840, Height: 1240},
	{Key: "library_logo", Width: 1280, Height: 720},
	{Key: "library_header_capsule", Width: 920, Height: 430},
	{Key: "event_cover", Width: 800, Height: 450},
	{Key: "event_header", Width: 1920, Height: 622},
}

// CropRect computes the crop rectangle from a focal point (focalX, focalY in [0,1])
// and a zoom factor for a given SteamAssetSpec. The returned rectangle fits inside
// masterBounds, has the target aspect ratio, and is centered on the focal point
// (clamped to bounds). zoom=1.0 is the default (largest crop fitting the aspect
// ratio); zoom=2.0 crops half the area, producing a 2× zoom-in effect.
func CropRect(masterBounds image.Rectangle, focalX, focalY, zoom float64, spec SteamAssetSpec) image.Rectangle {
	masterW := masterBounds.Dx()
	masterH := masterBounds.Dy()

	targetAR := float64(spec.Width) / float64(spec.Height)
	masterAR := float64(masterW) / float64(masterH)

	var cropW, cropH int
	if masterAR > targetAR {
		// master is wider than target: fit to height
		cropH = masterH
		cropW = int(math.Round(float64(masterH) * targetAR))
	} else {
		// master is taller than target: fit to width
		cropW = masterW
		cropH = int(math.Round(float64(masterW) / targetAR))
	}

	// Apply zoom: zoom > 1 shrinks the crop region (zoom in); zoom <= 0 treated as 1.
	if zoom > 0 && zoom != 1.0 {
		cropW = int(math.Round(float64(cropW) / zoom))
		cropH = int(math.Round(float64(cropH) / zoom))
	}

	// convert normalized focal point to absolute pixel coords
	focalAbsX := masterBounds.Min.X + int(math.Round(focalX*float64(masterW)))
	focalAbsY := masterBounds.Min.Y + int(math.Round(focalY*float64(masterH)))

	// center crop rect on focal point
	x0 := focalAbsX - cropW/2
	y0 := focalAbsY - cropH/2
	x1 := x0 + cropW
	y1 := y0 + cropH

	// clamp to master bounds
	if x0 < masterBounds.Min.X {
		x1 += masterBounds.Min.X - x0
		x0 = masterBounds.Min.X
	}
	if y0 < masterBounds.Min.Y {
		y1 += masterBounds.Min.Y - y0
		y0 = masterBounds.Min.Y
	}
	if x1 > masterBounds.Max.X {
		x0 -= x1 - masterBounds.Max.X
		x1 = masterBounds.Max.X
	}
	if y1 > masterBounds.Max.Y {
		y0 -= y1 - masterBounds.Max.Y
		y1 = masterBounds.Max.Y
	}

	return image.Rect(x0, y0, x1, y1)
}
