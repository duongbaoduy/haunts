package house

import (
  "glop/gui"
  "github.com/arbaal/mathgl"
  "math"
)

type HouseViewer struct {
  gui.Childless
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonFocuser
  gui.NonResponder
  gui.NonThinker

  house *houseDef

  zoom,angle float32
  mat,imat mathgl.Mat4
}

func MakeHouseViewer(house *houseDef, angle float32) *HouseViewer {
  var hv HouseViewer
  hv.EmbeddedWidget = &gui.BasicWidget{ CoreWidget: &hv }
  hv.Request_dims.Dx = 100
  hv.Request_dims.Dy = 100
  hv.Ex = true
  hv.Ey = true
  hv.house = house
  hv.angle = angle
  hv.Zoom(1)
  return &hv
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (hv *HouseViewer) Zoom(dz float64) {
  if dz == 0 {
    return
  }
  exp := math.Log(float64(hv.zoom)) + dz
  exp = float64(clamp(float32(exp), 2.5, 5.0))
  hv.zoom = float32(math.Exp(exp))
}

func (hv *HouseViewer) Drag(float64,float64) {
}

func (hv *HouseViewer) String() string {
  return "house viewer"
}

func (hv *HouseViewer) Draw(region gui.Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  
  var rooms rectObjectArray
  for _,room := range hv.house.Floors[0].Rooms {
    rooms = append(rooms, room)
  }
  rooms = rooms.Order()
  drawPrep()
  for i := len(rooms) - 1; i >= 0; i-- {
    room := rooms[i].(*Room)
    // TODO: Would be better to be able to just get the floor mats alone
    m_floor,_,m_left,_,m_right,_ := makeRoomMats(room.roomDef, region, 20-float32(room.X), 20-float32(room.Y), hv.angle, hv.zoom)

    drawWall(room.roomDef, m_floor, m_left, m_right, nil)
    drawFloor(room.roomDef, m_floor, nil)
    drawFurniture(m_floor, room.roomDef.Furniture, nil, 1)
    // drawWall(room *roomDef, wall *texture.Data, left, right mathgl.Mat4, temp *WallTexture)
  }
}














