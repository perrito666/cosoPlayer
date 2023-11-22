package main

func (s *SpriteStack) handle(actionID string, action func()) {
	for _, sprite := range s.sprites {
		if sprite.ID == actionID {
			sprite.action = action
			return
		}
	}
}
