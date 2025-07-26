package main

import (
	"slices"
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

func (c Categories) IsVisible(cat [2]string, showCats []string) bool {
	return cat[0][0] != '_' || slices.Contains(showCats, cat[0])
}

func NiceName(sound string, category [2]string, replaceWords map[string]string) string {
	out := strings.ReplaceAll(sound, "-", " ")
	for old, new := range replaceWords {
		out = strings.ReplaceAll(out, old, new)
	}
	out = cases.Title(language.AmericanEnglish, cases.NoLower).String(out)
	out = out[len(category[0]):]
	return out
}
