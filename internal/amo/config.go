package amo

import "ragbot/internal/util"

type ac struct {
	serviceName          string
	sourceFieldId        int
	sourceFieldValueId   int
	branchFieldId        int
	branchFieldValuesMap map[string]int
	interestFieldId      int
	summaryFieldId       int
	chatLinkFieldId      int
	tags                 []string
	dynamicTagsEnabled   bool
	keywordTagsMap       map[string][]string
}

var amoConfig *ac

func loadConfig() {
	amoConfig = &ac{
		serviceName:          util.GetEnvString("AMO_SERVICE_NAME", "RAG Ассистент"),
		sourceFieldId:        util.GetEnvInt("AMO_SOURCE_FIELD_ID", 0),
		sourceFieldValueId:   util.GetEnvInt("AMO_SOURCE_FIELD_VALUE_ID", 0),
		branchFieldId:        util.GetEnvInt("AMO_BRANCH_FIELD_ID", 0),
		branchFieldValuesMap: util.GetEnvStringIntMap("AMO_BRANCH_FIELD_VALUES_MAP", nil),
		interestFieldId:      util.GetEnvInt("AMO_INTEREST_FIELD_ID", 0),
		summaryFieldId:       util.GetEnvInt("AMO_SUMMARY_FIELD_ID", 0),
		chatLinkFieldId:      util.GetEnvInt("AMO_CHAT_LINK_FIELD_ID", 0),
		tags:                 util.GetEnvStringSlice("AMO_LEAD_TAGS", []string{}),
		dynamicTagsEnabled:   util.GetEnvBool("AMO_DYNAMIC_TAGS_ENABLED", true),
		keywordTagsMap:       loadKeywordTagsMap(),
	}
}

// loadKeywordTagsMap загружает карту ключевых слов и соответствующих тегов из переменной окружения
func loadKeywordTagsMap() map[string][]string {
	// Default keyword-tags mapping
	defaultMap := map[string][]string{
		"вопрос":        {"Вопрос"},
		"спросить":      {"Вопрос"},
		"консультация":  {"Консультация"},
		"консультирование": {"Консультация"},
		"консультацией": {"Консультация"},
		"консультации":  {"Консультация"},
		"информация":    {"Информация"},
		"информирование": {"Информация"},
		"помощь":        {"Помощь"},
		"поддержка":     {"Помощь"},
		"услуга":        {"Услуга"},
		"услуги":        {"Услуга"},
		"услуг":         {"Услуга"},
		"сервис":        {"Услуга"},
		"цена":          {"Цена"},
		"стоимость":     {"Цена"},
		"тариф":         {"Цена"},
		"запись":        {"Запись"},
		"записаться":    {"Запись"},
	}

	// Try to load from environment variable
	envMap := util.GetEnvJSON("AMO_KEYWORD_TAGS_MAP")
	if envMap != nil {
		result := make(map[string][]string)
		for keyword, tagsInterface := range envMap {
			if tagsArray, ok := tagsInterface.([]interface{}); ok {
				tags := make([]string, len(tagsArray))
				for i, tag := range tagsArray {
					if tagStr, ok := tag.(string); ok {
						tags[i] = tagStr
					}
				}
				result[keyword] = tags
			}
		}
		return result
	}

	return defaultMap
}
