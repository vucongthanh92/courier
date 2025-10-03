package localization

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type ResourceConfig struct {
	Lang      string
	Accept    string
	Resources []string
}

var localizer *i18n.Localizer
var bundle *i18n.Bundle

func InitResources(resources []string) error {
	bundle = i18n.NewBundle(language.Vietnamese)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, r := range resources {
		_, err := bundle.LoadMessageFile(r)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewLocalizer(resConfig ResourceConfig) {
	localizer = i18n.NewLocalizer(bundle, resConfig.Lang, resConfig.Accept)
}

func Localize(messageId string, templateData map[string]string) string {
	localizeConfig := i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: templateData,
	}

	if localizer == nil {
		localizer = i18n.NewLocalizer(bundle, "vi", "vi")
	}

	message, _ := localizer.Localize(&localizeConfig)
	if message == "" {
		return messageId
	}

	return message
}
