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
	}
}
