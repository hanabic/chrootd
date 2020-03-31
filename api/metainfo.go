package api

import "encoding/json"

type Metainfo struct {
	Id     []byte
	Name   string
	Rootfs string
}

func NewMetaFromBytes(bytes []byte) (*Metainfo, error) {
	r := &Metainfo{}
	if err := json.Unmarshal(bytes, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (m *Metainfo) Merge(o *Metainfo) {
	// TODO: Copy & overwrite as needed
	// TODO: add more until metadata struct definition is freezed
}

func (m *Metainfo) ToBytes() ([]byte, error) {
	r, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return r, nil
}
