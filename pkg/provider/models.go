package provider

type Model struct {
	name string
}

var (
	QWEN_35_FLASH        = Model{"qwen3.5-flash"}
	QWEN_36_FLASH        = Model{"qwen3.6-flash"}
	QWEN_35_PLUS         = Model{"qwen3.5-plus"}
	QWEN_36_PLUS         = Model{"qwen3.6-plus"}
	DEEPSEEK_V4_FLASH    = Model{"deepseek-v4-flash"}
	DEEPSEEK_V4_PRO      = Model{"deepseek-v4-pro"}
	GEMINI_31_FLASH_LITE = Model{"gemini-3.1-flash-lite"}
	GEMINI_3_FLASH       = Model{"gemini-3-flash-preview"}
	GEMINI_31_PRO        = Model{"gemini-3.1-pro-preview"}
)

func (m Model) String() string {
	return m.name
}
