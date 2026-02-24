package image

import (
	"image"
	"math"
	"testing"
)

func TestCropRect_FitsInsideMasterBounds(t *testing.T) {
	master := image.Rect(0, 0, 1920, 1080)
	for _, spec := range SteamAssets {
		rect := CropRect(master, 0.5, 0.5, 1.0, spec)
		if !rect.In(master) {
			t.Errorf("spec %s: crop rect %v exceeds master bounds %v", spec.Key, rect, master)
		}
	}
}

func TestCropRect_MatchesTargetAspectRatio(t *testing.T) {
	// Large square master so every target aspect ratio fits without clamping.
	master := image.Rect(0, 0, 5000, 5000)
	for _, spec := range SteamAssets {
		rect := CropRect(master, 0.5, 0.5, 1.0, spec)
		gotAR := float64(rect.Dx()) / float64(rect.Dy())
		wantAR := float64(spec.Width) / float64(spec.Height)
		if math.Abs(gotAR-wantAR) > 0.02 {
			t.Errorf("spec %s: aspect ratio %.4f, want %.4f", spec.Key, gotAR, wantAR)
		}
	}
}

func TestCropRect_TopLeftFocalPoint_ClampsToOrigin(t *testing.T) {
	master := image.Rect(0, 0, 1920, 1080)
	spec := SteamAssetSpec{Key: "test", Width: 920, Height: 430}

	rect := CropRect(master, 0.0, 0.0, 1.0, spec)

	if rect.Min != master.Min {
		t.Errorf("top-left at %v, want %v", rect.Min, master.Min)
	}
}

func TestCropRect_BottomRightFocalPoint_ClampsToCorner(t *testing.T) {
	master := image.Rect(0, 0, 1920, 1080)
	spec := SteamAssetSpec{Key: "test", Width: 920, Height: 430}

	rect := CropRect(master, 1.0, 1.0, 1.0, spec)

	if rect.Max != master.Max {
		t.Errorf("bottom-right at %v, want %v", rect.Max, master.Max)
	}
}
