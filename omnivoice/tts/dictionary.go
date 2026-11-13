package tts

import (
	"github.com/plexusone/elevenlabs-go/account"
	"github.com/plexusone/omnivoice-core/tts"
)

// DictionaryRules converts pronunciations into rules for an ElevenLabs
// pronunciation dictionary (account.CreatePronunciationDictionaryRequest).
// With phonemes true, a term with IPA becomes a phoneme rule; otherwise,
// or without IPA, a term with a respelling becomes an alias rule. Phoneme
// rules only take effect on models that honour phonemes; alias rules work
// with every model.
func DictionaryRules(ps []tts.Pronunciation, phonemes bool) account.PronunciationRules {
	var rules account.PronunciationRules
	for _, p := range ps {
		switch {
		case p.Term == "":
		case phonemes && p.IPA != "":
			rules = append(rules, account.PronunciationRule{Grapheme: p.Term, Phoneme: p.IPA})
		case p.Respelling != "":
			rules = append(rules, account.PronunciationRule{Grapheme: p.Term, Alias: p.Respelling})
		}
	}
	return rules
}
