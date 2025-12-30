package mp

// model type
const (
	ModelTypeLLM        = "llm"
	ModelTypeEmbedding  = "embedding"
	ModelTypeRerank     = "rerank"
	ModelTypeOcr        = "ocr"
	ModelTypeGui        = "gui"
	ModelTypePdfParser  = "pdf-parser"
	ModelTypeAsr        = "asr"
	ModelTypeText2Image = "text2image"
	//ModelTypeOcrDs      = "ocr-deepseek"
	//ModelTypeOcrPaddle  = "ocr-paddle"
)

// model provider
const (
	ProviderOpenAICompatible = "OpenAI-API-compatible"
	ProviderYuanJing         = "YuanJing"
	ProviderHuoshan          = "HuoShan"
	ProviderOllama           = "Ollama"
	ProviderQwen             = "Qwen"
	ProviderInfini           = "Infini"
	ProviderQianfan          = "QianFan"
	ProviderDeepSeek         = "DeepSeek"
)

var (
	_callbackUrl string
)

func Init(callbackUrl string) {
	if _callbackUrl != "" {
		panic("model provider already init")
	}
	_callbackUrl = callbackUrl
}
