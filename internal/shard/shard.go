package shard

func CreateShard() *Shard {
	return &Shard{
		Data: make(map[string]interface{}),
	}
}

func (s *Shard) Put(key string, value interface{}) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	s.Data[key] = value
}

func (s *Shard) Get(key string) (interface{}, bool) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	
	val, ok := s.Data[key]
	return val, ok
}