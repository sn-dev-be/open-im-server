package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	goI18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var (
	GlobalDefaultTr = &DefaultTranslator{
		localizes: make(map[Language]*goI18n.Localizer),
		jsonData:  make(map[Language]any),
	}
	Bundle = goI18n.NewBundle(language.English)
)

func init() {
	Bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
}

type DefaultTranslator struct {
	localizes map[Language]*goI18n.Localizer
	jsonData  map[Language]any
}

func (tr *DefaultTranslator) Dump(la Language) ([]byte, error) {
	return json.Marshal(tr.jsonData[la])
}

// TODO: improve multi-threading performance
func (tr *DefaultTranslator) Tr(la Language, key string) string {
	return tr.TrWithData(la, key, nil)
}

func (tr *DefaultTranslator) TrWithData(la Language, key string, templateData any) string {
	l, ok := tr.localizes[la]
	if !ok {
		l = tr.localizes[DefaultLanguage]
	}

	translation, err := l.Localize(&goI18n.LocalizeConfig{MessageID: key, TemplateData: templateData})
	if err != nil {
		// if _, tmpl, err := l.GetMessageTemplate(key, nil); err != nil {
		// 	return key
		// } else {
		// 	return tmpl.Other
		// }
	}

	return translation
}

func AddTranslator(translation []byte, language string) (err error) {
	_, err = Bundle.ParseMessageFileBytes(translation, language)
	if err != nil {
		return err
	}

	languageName := strings.TrimSuffix(language, filepath.Ext(language))

	GlobalDefaultTr.localizes[Language(languageName)] = goI18n.NewLocalizer(Bundle, languageName)

	j, err := yamlToJson(translation)
	if err != nil {
		return err
	}
	GlobalDefaultTr.jsonData[Language(languageName)] = j
	return
}

// TODO: singleton and multi-thread safe initialization
func NewDefaultTranslator(bundleDir string) (Translator, error) {
	stat, err := os.Stat(bundleDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", bundleDir)
	}

	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		return nil, err
	}

	for _, file := range entries {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".yaml" || file.Name() == "i18n.yaml" {
			continue
		}

		buf, err := os.ReadFile(filepath.Join(bundleDir, file.Name()))
		if err != nil {
			return nil, err
		}

		if _, err := Bundle.ParseMessageFileBytes(buf, file.Name()); err != nil {
			return nil, fmt.Errorf("parse language message file [%s] failed: %s", file.Name(), err)
		}

		languageName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		GlobalDefaultTr.localizes[Language(languageName)] = goI18n.NewLocalizer(Bundle, languageName)

		j, err := yamlToJson(buf)
		if err != nil {
			return nil, err
		}

		GlobalDefaultTr.jsonData[Language(languageName)] = j
	}
	return GlobalDefaultTr, nil
}
