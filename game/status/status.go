package status

import (
  "bytes"
  "encoding/gob"
  "encoding/json"
  "fmt"
)

type Kind string
const (
  Panic       Kind = "Panic"
  Terror      Kind = "Terror"
  Fire        Kind = "Fire"
  Poison      Kind = "Poison"
  Brutal      Kind = "Brutal"
  Unspecified Kind = "Unspecified"
)
type Primary int
const (
  Ego    Primary = iota
  Corpus
)
func (k Kind) Primary() Primary {
  switch k {
  case Panic: fallthrough
  case Terror:
    return Ego

  case Fire: fallthrough
  case Brutal: fallthrough
  case Poison:
    return Corpus
  }
  panic("Unknown status.Kind")
}

// Damage is something that affects a unit's current Ap or Hp.  A unit's Ap
// and Hp is affected through this mechanism so that Conditions have a chance
// to modify it before it actually gets applied.
type Damage struct {
  Dynamic
  Kind Kind
}

type Dynamic struct {
  Hp,Ap int
}

type Base struct {
  Ap_max int
  Hp_max int
  Corpus int
  Ego    int
  Sight  int
  Attack int
}

func MakeInst(b Base) Inst {
  var i Inst
  i.inst.Base = b
  return i
}

type inst struct {
  Base
  Dynamic
  Conditions []Condition
}

type Inst struct {
  // This prevents external code from modifying any data without goin through
  // the appropriate methods, but also allows us to provide accurate json and
  // gob methods.
  inst inst
}

func (s Inst) modifiedBase(kind Kind) Base {
  b := s.inst.Base
  for _,e := range s.inst.Conditions {
    b = e.ModifyBase(b, kind)
  }
  return b
}

func (s Inst) HpCur() int {
  return s.inst.Dynamic.Hp
}

func (s Inst) ApCur() int {
  return s.inst.Dynamic.Ap
}

func (s Inst) HpMax() int {
  hp_max := s.modifiedBase(Unspecified).Hp_max
  if hp_max < 0 { return 0 }
  return hp_max
}

func (s Inst) ApMax() int {
  ap_max := s.modifiedBase(Unspecified).Ap_max
  if ap_max < 0 { return 0 }
  return ap_max
}

func (s Inst) Corpus() int {
  corpus := s.modifiedBase(Unspecified).Corpus
  return corpus
}

func (s Inst) CorpusVs(kind Kind) int {
  corpus := s.modifiedBase(kind).Corpus
  return corpus
}

func (s Inst) Ego() int {
  ego := s.modifiedBase(Unspecified).Ego
  return ego
}

func (s Inst) EgoVs(kind Kind) int {
  ego := s.modifiedBase(kind).Ego
  return ego
}

// Calls either EgoVs or CorpusVs, depending on what Kind is specified.
// This function panics if neither corpus nor ego is specified.
func (s Inst) DefenseVs(kind Kind) int {
  switch kind.Primary() {
    case Corpus:
      return s.CorpusVs(kind)
    case Ego:
      return s.EgoVs(kind)
  }
  panic(fmt.Sprintf("Cannot call DefenseVs on kind '%v'", kind))
}

func (s Inst) AttackBonusWith(kind Kind) int {
  attack := s.modifiedBase(kind).Attack
  return attack
}

func (s Inst) Sight() int {
  sight := s.modifiedBase(Unspecified).Sight
  if sight < 0 { return 0 }
  return sight
}

func (s *Inst) ApplyCondition(c Condition) {
  for i := range s.inst.Conditions {
    if s.inst.Conditions[i].Kind() == c.Kind() && s.inst.Conditions[i].Buff() == c.Buff() {
      if s.inst.Conditions[i].Strength() <= c.Strength() {
        s.inst.Conditions[i] = c
      }

      // Regardless of whether it was displaced or not we don't need to keep
      // checking Conditions.  We can only have one condition of each type
      // and buff pair, and this one is it.
      return
    }
  }

  // If we didn't find an existing condition of this kind then we can safely
  // add it.
  s.inst.Conditions = append(s.inst.Conditions, c)
}

func (s *Inst) ApplyDamage(dap,dhp int, kind Kind) {
  dmg := Damage{ Dynamic: Dynamic{ Ap: dap, Hp: dhp }, Kind: kind }
  for _,c := range s.inst.Conditions {
    dmg = c.ModifyDamage(dmg)
  }
  s.inst.Ap += dmg.Ap
  s.inst.Hp += dmg.Hp
}

func (s *Inst) OnBegin() {
  s.inst.Hp = s.inst.Hp_max
  s.OnRound()
}

func (s *Inst) OnRound() {
  completed := make(map[Condition]bool)
  var dmgs []Damage
  for i := 0; i < len(s.inst.Conditions); i++ {
    dmg,done := s.inst.Conditions[i].OnRound()
    if dmg != nil {
      dmgs = append(dmgs, *dmg)
    }
    if done {
      completed[s.inst.Conditions[i]] = true
    }
  }

  s.inst.Ap = s.ApMax()
  for _,dmg := range dmgs {
    s.ApplyDamage(dmg.Ap, dmg.Hp, dmg.Kind)
  }

  // Negative Ap is as useless as zero, so just set it to zero for simplicity
  if s.inst.Ap < 0 {
    s.inst.Ap = 0
  }

  // Hp is capped at 0 as well, but also capped at its max
  if s.inst.Hp < 0 {
    s.inst.Hp = 0
  }
  if s.inst.Hp > s.HpMax() {
    s.inst.Hp = s.HpMax()
  }

  // Now remove all of the Conditions that completed
  num_complete := 0
  for i := 0; i < len(s.inst.Conditions); i++ {
    if completed[s.inst.Conditions[i]] {
      num_complete++
    } else {
      s.inst.Conditions[i - num_complete] = s.inst.Conditions[i]
    }
  }
  s.inst.Conditions = s.inst.Conditions[0 : len(s.inst.Conditions) - num_complete]
}

// Encoding routines - only support json and gob right now

func (si Inst) MarshalJSON() ([]byte, error) {
  return json.Marshal(si.inst)
}

func (si *Inst) UnmarshalJSON(data []byte) error {
  return json.Unmarshal(data, &si.inst)
}

func (si Inst) GobEncode() ([]byte, error) {
  buf := bytes.NewBuffer(nil)
  enc := gob.NewEncoder(buf)
  err := enc.Encode(si.inst)
  return buf.Bytes(), err
}

func (si *Inst) GobDecode(data []byte) error {
  dec := gob.NewDecoder(bytes.NewBuffer(data))
  return dec.Decode(&si.inst)
}
