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
	ID    string
	Title string
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
	return err
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
	row := s.db.QueryRow(`SELECT id, title FROM conversations WHERE id = ?`, id)
	var c Conversation
	if err := row.Scan(&c.ID, &c.Title); err != nil {
		return nil, err
	}
	return &c, nil
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
