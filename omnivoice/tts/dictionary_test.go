package tts

import (
	"testing"

	"github.com/plexusone/omnivoice-core/tts"
)

func TestDictionaryRules(t *testing.T) {
	ps := []tts.Pronunciation{
		{Term: "Miura", IPA: "ˈmjura", Respelling: "mee oora"},
		{Term: "Countach", IPA: "ˈkuntaʃ"},
		{Term: "Bertone"},
	}
	ipa := DictionaryRules(ps, true)
	if len(ipa) != 2 || ipa[0].Phoneme != "ˈmjura" || ipa[1].Phoneme != "ˈkuntaʃ" {
		t.Errorf("phoneme rules = %+v", ipa)
	}
	alias := DictionaryRules(ps, false)
	if len(alias) != 1 || alias[0].Grapheme != "Miura" || alias[0].Alias != "mee oora" {
		t.Errorf("alias rules = %+v", alias)
	}
	for _, r := range append(ipa, alias...) {
		if err := r.Validate(); err != nil {
			t.Errorf("%+v: %v", r, err)
		}
	}
}
