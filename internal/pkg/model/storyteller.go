package model

// Scene represents a single drawable scene (exactly one image per scene).
type StorytellerScene struct {
	ID         string `json:"id"`
	SceneIndex int    `json:"scene_index,omitempty"`
	Title      string `json:"title"`
	// 故事详情
	StoryDetails string `json:"story_details"`
	// 绘图结果
	DrawerOutput string `json:"drawer_output"`
}
