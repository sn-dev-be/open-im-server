package i18n

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenIMSDK/tools/log"
	fsI18n "github.com/openimsdk/open-im-server/v3/i18n"
	"gopkg.in/yaml.v3"
)

var GlobalTr Translator

type LangOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
	// Translation completion percentage
	Progress int `json:"progress"`
}

const DefaultLangOption = "Default"

var (
	LanguageOptions []*LangOption
)

func NewTranslator(c *I18n) (tr Translator, err error) {
	entries, err := fsI18n.Assets.ReadDir(".")
	if err != nil {
		// panic(err)
		return nil, err
	}

	for _, file := range entries {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".yaml" && file.Name() != "i18n.yaml" {
			continue
		}
		buf, err := fsI18n.Assets.ReadFile(file.Name())
		if err != nil {
			return nil, fmt.Errorf("read file failed: %s %s", file.Name(), err)
		}

		originalTr := struct {
			Msg map[string]map[string]interface{} `yaml:"msg"`
			Err map[string]interface{}            `yaml:"err"`
		}{}
		if err = yaml.Unmarshal(buf, &originalTr); err != nil {
			return nil, err
		}
		translation := make(map[string]interface{}, 0)
		for k, v := range originalTr.Msg {
			translation[k] = v
		}
		translation["msg"] = originalTr.Msg
		translation["err"] = originalTr.Err

		content, err := yaml.Marshal(translation)
		if err != nil {
			log.ZError(context.Background(), "marshal translation content failed", err, "fileName", file.Name())
			continue
		}

		if err = AddTranslator(content, file.Name()); err != nil {
			log.ZError(context.Background(), "add translator failed", err, "fileName", file.Name())
			continue
		}
	}
	GlobalTr = GlobalDefaultTr

	i18nFile, err := os.ReadFile(filepath.Join(c.BundleDir, "i18n.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read i18n file failed: %s", err)
	}

	s := struct {
		LangOption []*LangOption `yaml:"language_options"`
	}{}
	err = yaml.Unmarshal(i18nFile, &s)
	if err != nil {
		return nil, fmt.Errorf("i18n file parsing failed: %s", err)
	}
	LanguageOptions = s.LangOption
	for _, option := range LanguageOptions {
		option.Label = fmt.Sprintf("%s (%d%%)", option.Label, option.Progress)
	}
	return GlobalTr, err
}

func CheckLanguageIsValid(lang string) bool {
	if lang == DefaultLangOption {
		return true
	}
	for _, option := range LanguageOptions {
		if option.Value == lang {
			return true
		}
	}
	return false
}

// Tr use language to translate data. If this language translation is not available, return default english translation.
func Tr(lang Language, data string) string {
	if GlobalTr == nil {
		return data
	}
	translation := GlobalTr.Tr(lang, data)
	if translation == data {
		return GlobalTr.Tr(DefaultLanguage, data)
	}
	return translation
}

// TrWithData translate key with template data, it will replace the template data {{ .PlaceHolder }} in the translation.
func TrWithData(lang Language, key string, templateData any) string {
	if GlobalTr == nil {
		return key
	}
	translation := GlobalTr.TrWithData(lang, key, templateData)
	if translation == key {
		return GlobalTr.TrWithData(DefaultLanguage, key, templateData)
	}
	return translation
}
