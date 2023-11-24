package main

func (s *SpriteStack) register(actionID string, action func()) {
	if s.actionHandler == nil {
		s.actionHandler = map[string]func(){}
	}
	s.actionHandler[actionID] = action
}
