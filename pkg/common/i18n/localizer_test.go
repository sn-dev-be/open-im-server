package i18n_test

import (
	"os"
	"testing"

	"github.com/openimsdk/open-im-server/v3/pkg/common/i18n"
	"github.com/stretchr/testify/assert"
)

const i18nPath = "../../../i18n/"

func TestNewTranslator(t *testing.T) {
	translator, err := i18n.NewDefaultTranslator(i18nPath)
	assert.NoError(t, err)
	assert.Equal(t, translator.Tr(i18n.LanguageChinese, "msg.base.success"), "成功")
	assert.Equal(t, translator.Tr(i18n.LanguageEnglish, "msg.base.success"), "success")
}

func TestAddTranslator(t *testing.T) {
	enUS, err := os.ReadFile(i18nPath + "/en_US.yaml")
	assert.NoError(t, err)

	zhCN, err := os.ReadFile(i18nPath + "/zh_CN.yaml")
	assert.NoError(t, err)

	err = i18n.AddTranslator(enUS, "en_US.yaml")
	assert.NoError(t, err)

	err = i18n.AddTranslator(zhCN, "zh_CN.yaml")
	assert.NoError(t, err)

	assert.Equal(t, i18n.GlobalDefaultTr.Tr(i18n.LanguageChinese, "msg.base.success"), "成功")
	assert.Equal(t, i18n.GlobalDefaultTr.Tr(i18n.LanguageEnglish, "msg.base.success"), "success")
	t.Log(i18n.GlobalDefaultTr.TrWithData(i18n.LanguageChinese, "msg.base.test", map[string]interface{}{
		"Name": "hello",
	}))
}
