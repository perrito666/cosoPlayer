package main

func (s *SpriteStack) register(actionID string, action func()) {
	s.actionHandler[actionID] = action
}
