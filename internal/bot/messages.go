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
		"/start или /myid — получить свой chat_id\n" +
		"/update <id> <текст> — обновить фрагмент по ID\n" +
		"/delete <id> — удалить фрагмент по ID\n" +
		"/help — эта справка\n" +
		"\n" +
		"Все остальные сообщения будут интерпретированы как фрагменты для записи в базу знаний.\n" +
		"\n" +
		"**Как добавлять знания в базу:**\n" +
		"1. Вносите информацию маленькими фрагментами: небольшими простыми предложениями.\n" +
		"2. Начало и конец фрагмента должны быть осмысленными, в идеале должны совпадать с началом и концом предложения, а лучше абзаца, чтобы смысл содержался во фрагменте целиком.\n" +
		"3. Один фрагмент должен нести в себе одну «единицу смысла», одно понятие или описание. Не перегружайте фрагменты разной несвязанной друг с другом информацией.\n" +
		"4. Фрагменты должны перекрывать друг друга, чтобы ассистент мог собрать разные фрагменты в общую картину.\n" +
		"\n" +
		"**Например:**\n" +
		"абонемент это пропуск, дающий право посещения занятий в выбранных классах\n" +
		"абонементы бывают разных типов\n" +
		"абонемент типа Стандарт дает право посещения одного класса\n" +
		"абонемент типа Вездеход дает право посещения разных классов в одной студии\n" +
		"абонемент на месяц включает 8 или 12 занятий (в зависимости от типа абонемента)\n" +
		"при покупке нескольких абонементов действуют скидки\n" +
		"условия акций и скидок можно уточнить у администратора"
	msgAdminInvalidID         = "Неверный ID"
	msgAdminDeleteError       = "Ошибка удаления фрагмента #%d"
	msgAdminDeletedFormat     = "Удалён фрагмент #%d: %s"
	msgAdminUpdateUsage       = "Использование: /update <id> <новый текст>"
	msgAdminUpdateError       = "Ошибка обновления фрагмента #%d: %s"
	msgAdminUpdatedFormat     = "Обновлён фрагмент #%d: %s"
	msgAdminAddError          = "Ошибка добавления фрагмента: %s"
	msgAdminAdded             = "Добавлен фрагмент #%d: %s"
	msgAdminExists            = "Фрагмент уже существует: %s"
	msgUnknownCommand         = "Неизвестная команда"
	msgServiceUnavailable     = "Информация недоступна"
	msgInfoUnavailable        = "Информация недоступна"
	msgScheduleTitle          = "Расписание занятий:"
	msgPricesTitle            = "Цены на обучение:"
	msgAddressFormat          = "Студия «%s»: %s\n"
	msgPriceButtonFormat      = "Абонемент %s — %s₽"
	msgPriceDescriptionFormat = "*Абонемент %s*\n%s\n\n%s"
	msgPassHoursFormat        = "• Доступно %s занятий\n"
	msgGuestVisitsFormat      = "• Включает %s гостевых посещений для друзей\n"
	msgPassFreezeAllowed      = "• Разрешена «заморозка» на 30 дней\n"
	msgPassLifetimeFormat     = "• Срок действия: %s дн.\n"
	msgPriceFormat            = "• *Стоимость: %s₽*\n"
	msgScheduleLinkFormat     = "Расписание студии %s"
	msgStatsPrompt            = "Чтобы открыть статистику, нажмите кнопку:"
	msgStatsButton            = "Статистика"
	msgChatsPrompt            = "Чтобы открыть список чатов, нажмите кнопку:"
	msgChatsButton            = "Чаты"
	msgChannelPrompt          = "Чтобы открыть телеграм-канал ШТБП, нажмите кнопку:"
)

const (
	promptUserPrefix        = "Пользователь: "
	promptAssistantPrefix   = "Помощник: "
	promptSummarizeGist     = "Суммаризируй диалог пользователя в двух предложениях с упоминанием выбранных пользователем танцевальных направлений (если таковые были), а также выбранного филиала (если он был выбран):\n%s\nРезюме:"
	promptSummarizeTitle    = "Сократи суть обращения до заголовка из 5-6 слов:\n%s\nСуть:"
	promptSummarizeInterest = "Если пользователь запрашивал информацию о конкретных танцевальных классах или направлениях, перечисли их. Если таковых нет, верни пустую строку:\n%s\nПользователь интересовался классами:"
	historyCallRequested    = "** хочет, чтобы ему перезвонили **"
	historyConfirmYes       = "** подтвердил контактные данные **"
	historyConfirmNo        = "** опроверг контактные данные **"
)
