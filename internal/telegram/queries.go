package telegram

var storeChatQuery string = `
	INSERT INTO chats
	VALUES ($1, $2, $3)
	ON CONFLICT DO NOTHING;`

var getAllQuery string = `
	SELECT id
	FROM chats;`
