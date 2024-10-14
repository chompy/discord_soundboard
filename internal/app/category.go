package app

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Categories [][2]string

func (c Categories) SoundInCategory(sound string, cat [2]string) bool {
	for _, icat := range c {
		if strings.HasPrefix(sound, icat[0]) {
			return cat[0] == icat[0]
		}
	}
	return false
}

func NiceName(sound string) string {
	out := strings.ReplaceAll(sound, "-", " ")
	out = cases.Title(language.AmericanEnglish, cases.NoLower).String(out)
	return out
}
