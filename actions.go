package main

func (s *SpriteStack) register(actionID string, action func() error) {
	if s.actionHandler == nil {
		s.actionHandler = map[string]func() error{}
	}
	s.actionHandler[actionID] = action
}
