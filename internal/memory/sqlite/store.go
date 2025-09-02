package sqlite

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Message struct {
	ID             int64
	ConversationID string
	Role           string
	Content        string
}

type Conversation struct {
	ID                  string
	Title               string
	ContextPromptTokens int
	ContextAnswerTokens int
	PromptMessageCount  int
	AnswerMessageCount  int
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	st := &Store{db: db}
	if err := st.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return st, nil
}

func (s *Store) init() error {
	// Base tables (no columns that might fail on old DBs)
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		title TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id TEXT,
		role TEXT,
		content TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return err
	}
	// Migrations: add context columns if missing
	if err := s.ensureColumn("conversations", "context_prompt_tokens", "INTEGER", "0"); err != nil {
		return err
	}
	if err := s.ensureColumn("conversations", "context_answer_tokens", "INTEGER", "0"); err != nil {
		return err
	}
	if err := s.ensureColumn("conversations", "prompt_message_count", "INTEGER", "0"); err != nil {
		return err
	}
	if err := s.ensureColumn("conversations", "answer_message_count", "INTEGER", "0"); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensureColumn(table, column, colType, defaultVal string) error {
	ok, err := s.hasColumn(table, column)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	_, err = s.db.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + colType + " DEFAULT " + defaultVal)
	return err
}

func (s *Store) hasColumn(table, column string) (bool, error) {
	rows, err := s.db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	var (
		cid         int
		name, ctype string
		notnull, pk int
		deflt       sql.NullString
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &deflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) CreateOrGetConversation(id string, title string) (*Conversation, error) {
	if id == "" {
		return nil, errors.New("conversation id required")
	}
	_, err := s.db.Exec(`INSERT OR IGNORE INTO conversations(id, title) VALUES(?, ?)`, id, title)
	if err != nil {
		return nil, err
	}
	row := s.db.QueryRow(`SELECT id, title, context_prompt_tokens, context_answer_tokens, prompt_message_count, answer_message_count FROM conversations WHERE id = ?`, id)
	var c Conversation
	if err := row.Scan(&c.ID, &c.Title, &c.ContextPromptTokens, &c.ContextAnswerTokens, &c.PromptMessageCount, &c.AnswerMessageCount); err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) UpdateContextUsage(conversationID string, promptTokens, answerTokens int) error {
	_, err := s.db.Exec(`UPDATE conversations SET context_prompt_tokens = ?, context_answer_tokens = ? WHERE id = ?`, promptTokens, answerTokens, conversationID)
	return err
}

func (s *Store) UpdateContextStats(conversationID string, promptTokens, answerTokens, promptCount, answerCount int) error {
	_, err := s.db.Exec(`UPDATE conversations SET context_prompt_tokens = ?, context_answer_tokens = ?, prompt_message_count = ?, answer_message_count = ? WHERE id = ?`, promptTokens, answerTokens, promptCount, answerCount, conversationID)
	return err
}

func (s *Store) AppendMessage(conversationID, role, content string) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO messages(conversation_id, role, content) VALUES(?, ?, ?)`, conversationID, role, content)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListMessages(conversationID string, limit int) ([]Message, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT id, conversation_id, role, content FROM messages WHERE conversation_id = ? ORDER BY id ASC LIMIT ?`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.Role, &m.Content); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) ClearConversation(conversationID string) error {
	_, err := s.db.Exec(`DELETE FROM messages WHERE conversation_id = ?`, conversationID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`UPDATE conversations SET context_prompt_tokens = 0, context_answer_tokens = 0, prompt_message_count = 0, answer_message_count = 0 WHERE id = ?`, conversationID)
	return err
}
