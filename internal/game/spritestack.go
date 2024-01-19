package game

import (
	"image/color"
	"rotate-test/internal/res"

	"github.com/hajimehoshi/ebiten/v2"
)

type SpriteStack struct {
	RVec2
	*res.ImageSheet
	LayerDistance float64
	z             float64
}

func NewSpriteStackFromImageSheet(imagesheet *res.ImageSheet) *SpriteStack {
	return &SpriteStack{
		ImageSheet:    imagesheet,
		LayerDistance: 1,
		z:             1,
	}
}

func (s *SpriteStack) Clone() *SpriteStack {
	return &SpriteStack{
		RVec2:         s.RVec2,
		ImageSheet:    s.ImageSheet,
		z:             s.z,
		LayerDistance: s.LayerDistance,
	}
}

func (s *SpriteStack) Z() float64 {
	return s.z
}

func (s *SpriteStack) SetZ(z float64) {
	s.z = z
}

func (s *SpriteStack) Draw(drawOpts DrawOpts) {
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterLinear
	// Rotate about center.
	op.GeoM.Translate(-s.HalfWidth(), -s.HalfHeight())
	op.GeoM.Rotate(s.Angle())
	op.GeoM.Translate(s.HalfWidth(), s.HalfHeight())
	// Translate to position.
	op.GeoM.Translate(s.X(), s.Y())
	op.GeoM.Concat(drawOpts.GeoM)
	for col := 0; col < s.ImageSheet.Cols(); col++ {
		op.ColorScale.Reset()
		r := float64(col) / float64(s.ImageSheet.Cols())
		c := uint8(150.0 + 105*r)
		op.ColorScale.ScaleWithColor(color.NRGBA{c, c, c, 255})
		drawOpts.Image.DrawImage(s.ImageSheet.At(col, 0), op)
		op.GeoM.Translate(0, -s.LayerDistance*drawOpts.Z)
	}
}

func (s *SpriteStack) Position() Vec2 {
	return Vec2{s.X(), s.Y()}
}

func (s *SpriteStack) Size() Vec2 {
	return Vec2{float64(s.Bounds().Dx()), float64(s.Bounds().Dy())}
}
