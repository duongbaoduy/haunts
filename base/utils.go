package base

import (
  "os"
  "encoding/json"
  "encoding/gob"
  "io/ioutil"
  "strings"
  "path/filepath"
  "image/color"
  "gl"
)

// Opens the file named by path, reads it all, decodes it as json into target,
// then closes the file.  Returns the first error found while doing this or nil.
func LoadJson(path string, target interface{}) error {
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  defer f.Close()
  data, err := ioutil.ReadAll(f)
  if err != nil {
    return err
  }
  err = json.Unmarshal(data, target)
  return err
}

func SaveJson(path string, source interface{}) error {
  data, err := json.Marshal(source)
  if err != nil {
    return err
  }
  f, err := os.Create(path)
  if err != nil {
    return err
  }
  defer f.Close()
  _,err = f.Write(data)
  return err
}

// Opens the file named by path, reads it all, decodes it as gob into target,
// then closes the file.  Returns the first error found while doing this or nil.
func LoadGob(path string, target interface{}) error {
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  defer f.Close()
  dec := gob.NewDecoder(f)
  err = dec.Decode(target)
  return err
}

func SaveGob(path string, source interface{}) error {
  f, err := os.Create(path)
  if err != nil {
    return err
  }
  defer f.Close()
  enc := gob.NewEncoder(f)
  err = enc.Encode(source)
  return err
}

// Returns a path rel such that filepath.Join(a, rel) and b refer to the same
// file.  a and b must both be relative paths or both be absolute paths.  If
// they are not then b will be returned in either case.
func RelativePath(a,b string) string {
  if filepath.IsAbs(a) != filepath.IsAbs(b) {
    return b
  }
  aparts := strings.Split(filepath.ToSlash(filepath.Clean(a)), "/")
  bparts := strings.Split(filepath.ToSlash(filepath.Clean(b)), "/")
  for len(aparts) > 0 && len(bparts) > 0 && aparts[0] == bparts[0] {
    aparts = aparts[1:]
    bparts = bparts[1:]
  }
  for i := range aparts {
    aparts[i] = ".."
  }
  ret := filepath.Join(filepath.Join(aparts...), filepath.Join(bparts...))
  return filepath.Clean(ret)
}

var datadir string
func SetDatadir(_datadir string) {
  datadir = _datadir
}
func GetStoreVal(key string) string {
  var store map[string]string
  LoadJson(filepath.Join(datadir, "store"), &store)
  if store == nil {
    store = make(map[string]string)
  }
  val := store[key]
  return val
}

func SetStoreVal(key,val string) {
  var store map[string]string
  path := filepath.Join(datadir, "store")
  LoadJson(path, &store)
  if store == nil {
    store = make(map[string]string)
  }
  store[key] = val
  SaveJson(path, store)
}

type ColorStack struct {
  colors []color.NRGBA
}
func (cs *ColorStack) Push(r,g,b,a float64) {
  c := color.NRGBA{ byte(255 * r), byte(255 * g), byte(255 * b), byte(255 * a) }
  cs.colors = append(cs.colors, c)
}
func (cs *ColorStack) Pop() {
  cs.colors = cs.colors[0 : len(cs.colors) - 1]
}
func (cs *ColorStack) subApply(n int) (r,g,b,a float64) {
  if n < 0 {
    return 1, 1, 1, 0
  }
  dr, dg, db, da := cs.subApply(n - 1)
  a = float64(cs.colors[n].A) / 255
  r = float64(cs.colors[n].R) / 255 * a + dr * (1 - a)
  g = float64(cs.colors[n].G) / 255 * a + dg * (1 - a)
  b = float64(cs.colors[n].B) / 255 * a + db * (1 - a)
  a = a + (1 - a) * da
  return
}
func (cs *ColorStack) Apply() {
  gl.Color4d(cs.subApply(len(cs.colors) - 1))
}
func (cs *ColorStack) ApplyWithAlpha(alpha float64) {
  r,g,b,a := cs.subApply(len(cs.colors) - 1)
  gl.Color4d(r, g, b, a * alpha)
}

