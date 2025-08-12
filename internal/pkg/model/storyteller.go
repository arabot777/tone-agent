package model

// StorytellerOutput mirrors the strict JSON schema returned by the storyteller agent.
// It is intentionally stable to keep prompt-schema and code in lockstep.
type StorytellerOutput struct {
	Locale       string  `json:"locale"`
	DeferDrawing bool    `json:"defer_drawing"`
	Title        string  `json:"title"`
	Scenes       []Scene `json:"scenes"`
}

// Scene represents a single drawable scene (exactly one image per scene).
type Scene struct {
	ID         string      `json:"id"`
	SceneIndex int         `json:"scene_index,omitempty"`
	Title      string      `json:"title"`
	Narrative  string      `json:"narrative"`
	VisualBrief VisualBrief `json:"visual_brief"`
	DrawInput  string      `json:"draw_input"`
	Style      string      `json:"style,omitempty"`
	Priority   int         `json:"priority,omitempty"`
}

// VisualBrief is stored as structured fields but may be partially filled.
type VisualBrief struct {
	Characters  any `json:"characters"`
	Setting     any `json:"setting"`
	Composition any `json:"composition"`
	Mood        any `json:"mood"`
	KeyDetails  any `json:"key_details"`
}

// Validate performs minimal checks for downstream logic.
func (o *StorytellerOutput) Validate() bool {
	if o == nil {
		return false
	}
	if len(o.Scenes) == 0 {
		return false
	}
	for _, s := range o.Scenes {
		if s.DrawInput == "" {
			return false
		}
	}
	return true
}
