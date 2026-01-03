package ai

// AIAnalyzer defines the interface for AI analysis services.
type AIAnalyzer interface {
	Analyze(report string, summarize bool) (string, error)
}
