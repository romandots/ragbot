package conversation

import (
	"sync"
)

// HistoryItem — одна запись истории (пользователь или бот).
type HistoryItem struct {
	Role    string // "user" или "assistant"
	Content string
}

// historyStore хранит карты chatID → []HistoryItem
var (
	historyStore = make(map[int64][]HistoryItem)
	historyMu    sync.Mutex
)

// AppendHistory добавляет новый элемент в историю конкретного чата.
func AppendHistory(chatID int64, role, text string) {
	historyMu.Lock()
	defer historyMu.Unlock()
	historyStore[chatID] = append(historyStore[chatID], HistoryItem{Role: role, Content: text})
	// Если хочется ограничить длину истории (например, максимум 20 сообщений), можно:
	if len(historyStore[chatID]) > 20 {
		historyStore[chatID] = historyStore[chatID][len(historyStore[chatID])-20:]
	}
}

// GetHistory возвращает копию истории чата.
func GetHistory(chatID int64) []HistoryItem {
	historyMu.Lock()
	defer historyMu.Unlock()
	// Копируем срез, чтобы не дать внешнему коду править исходный
	items := historyStore[chatID]
	result := make([]HistoryItem, len(items))
	copy(result, items)
	return result
}
