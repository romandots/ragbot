package bot

const (
	msgCallManagerButton    = "Хочу, чтобы мне перезвонили"
	msgCallManagerPrompt    = "Чтобы продолжить общение с нашим менеджером, нажмите кнопку:"
	msgConfirmYes           = "Да"
	msgConfirmNo            = "Нет"
	msgConfirmContactFormat = "Мы нашли ваши контактные данные: %s, %s. Всё верно?"
	msgAskName              = "Как к вам можно обращаться?"
	msgAskPhone             = "Напишите ваш телефон для связи."
	msgManagerWillCall      = "Наш менеджер свяжется с вами в ближайшее время."
	msgAdminSummaryFormat   = "%s (%s): %s\n\n%s"
	msgAdminErrorFormat     = "Возникла ошибка: %s"
	msgUserError            = "Возникла ошибка. Пожалуйста, попробуйте повторить ваш запрос позднее."
	msgAdminMyIDFormat      = "Ваш CHAT ID: %d"
	msgAdminHelp            = "Команды администратора:\n" +
		"/help — эта справка\n" +
		"/myid — получить свой chat_id\n" +
		"/delete <id> — удалить фрагмент по ID\n" +
		"/update <id> <текст> — обновить фрагмент по ID\n\n" +
		"Все остальное будет интерпретировано как запись в базу знаний"
	msgAdminInvalidID     = "Неверный ID"
	msgAdminDeleteError   = "Ошибка удаления"
	msgAdminDeletedFormat = "Удалён фрагмент %d"
	msgAdminUpdateUsage   = "Использование: /update <id> <новый текст>"
	msgAdminUpdateError   = "Ошибка обновления"
	msgAdminUpdatedFormat = "Обновлён фрагмент %d"
	msgAdminAddError      = "Ошибка добавления"
	msgAdminAdded         = "Добавлено"
	msgAdminExists        = "Уже существует"
)

const (
	promptUserPrefix        = "Пользователь: "
	promptAssistantPrefix   = "Помощник: "
	promptSummarizeGist     = "Суммаризируй диалог пользователя в двух предложениях:\n%s\nРезюме:"
	promptSummarizeTitle    = "Суммаризируй суть запроса в 3-5 словах, кроме того, что пользователь запросил звонок:\n%s\nСуть:"
	promptSummarizeInterest = "Вычлени, если есть, упоминания танцевальных классов или направлений, выбранных пользователем. Если таковых нет, верни пустую строку:\n%s\nКлассы:"
	historyCallRequested    = "** хочет, чтобы ему перезвонили **"
	historyConfirmYes       = "** подтвердил контактные данные **"
	historyConfirmNo        = "** опроверг контактные данные **"
)
