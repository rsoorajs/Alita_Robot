package modules

import (
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

// antispamModule provides anti-spam logic for group chats.
//
// It tracks spam levels per chat and enforces limits to prevent spammy behavior.
var antispamModule = moduleStruct{
	moduleName: "antispam",
	antiSpam:   map[int64]*antiSpamInfo{},
}

// checkSpammed checks if a chat has exceeded any anti-spam levels.
//
// Returns true if any spam threshold is exceeded, otherwise false.
func (moduleStruct) checkSpammed(chatId int64, levels []antiSpamLevel) bool {
	_asInfo, ok := antispamModule.antiSpam[chatId]
	if !ok {
		// Assign a new AntiSpamInfo to the chatId because not found
		antispamModule.antiSpam[chatId] = &antiSpamInfo{
			Levels: levels,
		}
		return false
	}
	newLevels := make([]antiSpamLevel, len(_asInfo.Levels))
	var spammed bool
	for n, level := range _asInfo.Levels {
		// Expire the _asInfo if current time becomes greater than expiration time
		if level.CurrTime+level.Expiry <= time.Duration(time.Now().UnixNano()) {
			// Allocate a new 'current time' with count reset to 0 if expired
			level.CurrTime = time.Duration(time.Now().UnixNano())
			level.Count = 0
			level.Spammed = false
		}
		level.Count += 1
		if (level.Count + 1) > level.Limit {
			// fmt.Println("level", n, "has been spammed with count", level.Count, "while the limit was", level.Limit)
			level.Spammed = true
		}
		newLevels[n] = level
		if !spammed && level.Spammed {
			spammed = true
		}
	}
	_asInfo.Levels = newLevels
	antispamModule.antiSpam[chatId] = _asInfo
	return spammed
}

// spamCheck applies the default anti-spam check for a chat.
//
// Returns true if the chat is currently spamming, otherwise false.
func (moduleStruct) spamCheck(chatId int64) bool {
	// if sql.IsUserSudo(chatId) {
	//	return false
	// }
	curr := time.Duration(time.Now().UnixNano())
	return antispamModule.checkSpammed(chatId, []antiSpamLevel{
		{
			CurrTime: curr,
			Limit:    18,
			Expiry:   time.Second,
		},
	})
}

// LoadAntispam registers the anti-spam message handler with the dispatcher.
//
// This handler checks for spam on every message and ends the handler group if spam is detected.
func LoadAntispam(dispatcher *ext.Dispatcher) {
	dispatcher.AddHandlerToGroup(
		handlers.NewMessage(
			message.All,
			func(bot *gotgbot.Bot, ctx *ext.Context) error {
				if antispamModule.spamCheck(ctx.EffectiveChat.Id) {
					return ext.EndGroups
				}
				return ext.ContinueGroups
			},
		), -2,
	)
}
