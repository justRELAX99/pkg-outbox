package entity

import (
	"encoding/json"
)

type ServicesList struct {
	ServicesMap map[string]string `json:"services"`
}

func (s *ServicesList) UnmarshalJSON(b []byte) error {
	s.ServicesMap = make(map[string]string)
	tmp := make(map[string]string)
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	for key, val := range tmp {
		s.ServicesMap[key] = val
	}
	return nil
}

func (s *ServicesList) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.ServicesMap)
}

func (s ServicesList) Get(key string) (string, bool) {
	val, ok := s.ServicesMap[key]
	return val, ok
}
