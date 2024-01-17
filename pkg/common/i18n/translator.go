package i18n

import "strings"

const (
	LanguageChinese            Language = "zh_CN"
	LanguageChineseTraditional Language = "zh_TW"
	LanguageEnglish            Language = "en_US"
	LanguageGerman             Language = "de_DE"
	LanguageFrench             Language = "fr_FR"
	LanguageJapanese           Language = "ja_JP"
	LanguageKorean             Language = "ko_KR"

	DefaultLanguage = LanguageEnglish
)

type Language string

func (t Language) Abbr() string {
	s := string(t)
	if idx := strings.Index(s, "_"); idx > 0 {
		return s[0:idx]
	}
	return s
}

type Translator interface {
	Tr(lang Language, key string) string
	TrWithData(lang Language, key string, templateData any) string
	Dump(lang Language) ([]byte, error)
}
