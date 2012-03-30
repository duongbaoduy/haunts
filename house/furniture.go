package house

import (
  "github.com/runningwild/haunts/base"
  "github.com/runningwild/haunts/texture"
  "github.com/runningwild/mathgl"
)

func MakeFurniture(name string) *Furniture {
  f := Furniture{ Defname: name }
  base.GetObject("furniture", &f)
  return &f
}

func GetAllFurnitureNames() []string {
  return base.GetAllNamesInRegistry("furniture")
}

func LoadAllFurnitureInDir(dir string) {
  base.RemoveRegistry("furniture")
  base.RegisterRegistry("furniture", make(map[string]*furnitureDef))
  base.RegisterAllObjectsInDir("furniture", dir, ".json", "json")
}

type Furniture struct {
  Defname string
  *furnitureDef

  // Position of this object in board coordinates.
  X,Y int

  // Index into furnitureDef.Texture_paths
  Rotation int
}

// Changes the position of this object such that it fits within the specified
// dimensions, if possible
func (f *Furniture) Constrain(dx,dy int) {
  cdx,cdy := f.Dims()
  if f.X + cdx > dx {
    f.X += dx - f.X + cdx
  }
  if f.Y + cdy > dy {
    f.Y += dy - f.Y + cdy
  }
}

func (f *Furniture) Pos() (int, int) {
  return f.X, f.Y
}

func (f *Furniture) RotateLeft() {
  f.Rotation = (f.Rotation + 1) % len(f.Orientations)
}

func (f *Furniture) RotateRight() {
  f.Rotation = (f.Rotation - 1 + len(f.Orientations)) % len(f.Orientations)
}

type furnitureOrientation struct {
  Dx,Dy int
  Texture texture.Object `registry:"autoload"`
}

// All instances of the same piece of furniture have this data in common
type furnitureDef struct {
  // Name of the object - should be unique among all furniture
  Name string

  // All available orientations for this piece of furniture
  Orientations []furnitureOrientation

  // Whether or not this piece of furniture blocks line-of-sight.  If a piece
  // of furniture blocks los, then the entire piece blocks los, regardless of
  // orientation.
  Blocks_los bool
}

func (f *Furniture) Dims() (int, int) {
  orientation := f.Orientations[f.Rotation]
  return orientation.Dx, orientation.Dy
}

func (f *Furniture) Render(pos mathgl.Vec2, width float32) {
  orientation := f.Orientations[f.Rotation]
  dy := width * float32(orientation.Texture.Data().Dy()) / float32(orientation.Texture.Data().Dx())
  orientation.Texture.Data().Render(float64(pos.X), float64(pos.Y), float64(width), float64(dy))
}
