package sync

// Export transformer types for black-box testing.
var NewTrimSpaceTransformer = newTrimSpaceTransformer
var NewQuoteTransformer = newQuoteTransformer
var NewChainTransformer = newChainTransformer

func newTrimSpaceTransformer() *TrimSpaceTransformer { return &TrimSpaceTransformer{} }
func newQuoteTransformer() *QuoteTransformer         { return &QuoteTransformer{} }
func newChainTransformer(ts ...ValueTransformer) *ChainTransformer {
	return NewChainTransformer(ts...)
}
